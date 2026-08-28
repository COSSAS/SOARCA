// Package fin implements the HTTP/JSON handlers for the new pull-based Fin
// protocol (see docs/adr/FIN-WEBHOOK-PROTOCOL-PROPOSAL.md): registration,
// long-poll job claiming, result submission, status-ping lease renewal,
// unregistration, and read-only discovery.
package fin

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	"soarca/internal/logger"
	apiError "soarca/pkg/api/error"
	"soarca/pkg/core/capability/fin/queue"
	"soarca/pkg/core/capability/fin/token"
	"soarca/pkg/models/fin"
	"soarca/pkg/utils/guid"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Empty struct{}

var log *logger.Log

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// finContextKey is the gin.Context key RequireFinToken stores the
// authenticated fin.Record under, for handlers to retrieve.
const finContextKey = "fin_record"

// IFinRepository is the subset of finrepository.IFinRepository this
// package depends on (avoids an import of internal/database/fin from
// pkg/..., which must not depend on internal/...).
type IFinRepository interface {
	Register(record fin.Record) error
	Get(finId string) (fin.Record, error)
	FindByTokenHash(tokenHash string) (fin.Record, error)
	List() ([]fin.Record, error)
	Touch(finId string, lastSeen time.Time) error
	Unregister(finId string) error
}

// Config holds the server-chosen operational defaults handed back to a Fin
// at registration (RegisterResponse), plus the shared secret gating new
// registrations.
type Config struct {
	// RegistrationToken gates POST /fin/register. An empty value means
	// registration is not configured at all - Register then always fails
	// closed (rather than silently accepting any registration attempt).
	RegistrationToken      string
	PollIntervalSeconds    int
	LongPollTimeoutSeconds int
	JobLeaseSeconds        int
	// StaleAfterSeconds is the threshold (in seconds, since LastSeen) after
	// which List/Get mark a registered Fin as Stale in their response, for
	// GUI/operator visibility. It should match the threshold used by
	// pkg/core/capability/fin.Capability's fail-fast liveness check so the
	// two stay consistent; a value <= 0 falls back to defaultStaleAfter.
	StaleAfterSeconds int
}

const defaultStaleAfter = 2 * time.Minute

type FinHandler struct {
	repository IFinRepository
	queue      *queue.Queue
	config     Config
	guid       guid.IGuid
}

func NewFinHandler(repository IFinRepository, jobQueue *queue.Queue, config Config, guid guid.IGuid) *FinHandler {
	return &FinHandler{repository: repository, queue: jobQueue, config: config, guid: guid}
}

// ############################################################################
// Registration-gated (POST /fin/register)
// ############################################################################

// Register
//
//	@Summary	register a new Fin and obtain its fin_token
//	@Schemes
//	@Description	register a new Fin process, declaring the capability types it can execute, and obtain its fin_id/fin_token
//	@Tags			fin
//	@Accept			json
//	@Produce		json
//	@Param			data	body		fin.RegisterRequest	true	"registration"
//	@Success		201		{object}	fin.RegisterResponse
//	@failure		400		{object}	api.Error
//	@failure		403		{object}	api.Error
//	@Router			/fin/register [POST]
func (finHandler *FinHandler) Register(g *gin.Context) {
	const route = "POST /fin/register"

	var request fin.RegisterRequest
	if err := g.ShouldBindJSON(&request); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse registration request", route, err.Error())
		return
	}

	if finHandler.config.RegistrationToken == "" {
		log.Warning("rejecting fin registration attempt: no FIN_REGISTRATION_TOKEN is configured")
		apiError.SendErrorResponse(g, http.StatusServiceUnavailable, "Fin registration is not configured", route, "")
		return
	}
	if !token.Equal(request.RegistrationToken, finHandler.config.RegistrationToken) {
		err := fin.ErrRegistrationTokenInvalid{}
		log.Warning(err)
		apiError.SendErrorResponse(g, http.StatusForbidden, err.Error(), route, "")
		return
	}
	if len(request.Capabilities) == 0 {
		apiError.SendErrorResponse(g, http.StatusBadRequest, "At least one capability is required", route, "")
		return
	}
	for _, capability := range request.Capabilities {
		if capability.Type == "" {
			apiError.SendErrorResponse(g, http.StatusBadRequest, "Every capability requires a non-empty type", route, "")
			return
		}
	}

	finToken, err := token.Generate()
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError, "Failed to generate fin token", route, "")
		return
	}

	record := fin.Record{
		FinId:           finHandler.guid.New().String(),
		FinTokenHash:    token.Hash(finToken),
		DisplayName:     request.DisplayName,
		ProtocolVersion: request.ProtocolVersion,
		Capabilities:    request.Capabilities,
		RegisteredAt:    time.Now(),
		LastSeen:        time.Now(),
	}

	if err := finHandler.repository.Register(record); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError, "Failed to register fin", route, "")
		return
	}

	log.Info("registered fin ", record.FinId, " (", record.DisplayName, ") with capabilities ", capabilityTypes(record.Capabilities))

	g.JSON(http.StatusCreated, fin.RegisterResponse{
		FinId:                  record.FinId,
		FinToken:               finToken,
		PollIntervalSeconds:    finHandler.config.PollIntervalSeconds,
		LongPollTimeoutSeconds: finHandler.config.LongPollTimeoutSeconds,
		JobLeaseSeconds:        finHandler.config.JobLeaseSeconds,
	})
}

// ############################################################################
// Fin-token-gated (poll / job result / status ping / unregister)
// ############################################################################

// RequireFinToken authenticates a call by its Authorization: Bearer
// fin_token header, resolving it to the Fin that presented it. On success,
// the resolved fin.Record is stored in the gin context for the handler to
// retrieve. Routes that use this middleware must be registered with it via
// gin's route-group Use, not the global engine, since it must not apply to
// Register (which has no fin_token yet) or the admin/dashboard read routes
// (List/Get, which are not Fin-authenticated calls at all).
func (finHandler *FinHandler) RequireFinToken(g *gin.Context) {
	const route = "fin bearer auth"

	presentedToken, ok := bearerToken(g)
	if !ok {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "Missing or malformed Authorization header", route, "")
		g.Abort()
		return
	}

	record, err := finHandler.repository.FindByTokenHash(token.Hash(presentedToken))
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "Invalid or unknown fin token", route, "")
		g.Abort()
		return
	}

	g.Set(finContextKey, record)
	g.Next()
}

// Poll
//
//	@Summary	long-poll for the next job matching this Fin's registered capability types
//	@Schemes
//	@Description	long-poll for the next job matching this Fin's registered capability types. Returns 204 No Content if long_poll_timeout_seconds elapses with no job available - callers should simply poll again.
//	@Tags			fin
//	@Accept			json
//	@Produce		json
//	@Param			data	body		fin.PollRequest	false	"poll hints"
//	@Success		200		{object}	fin.PollResponse
//	@Success		204
//	@failure		401		{object}	api.Error
//	@Router			/fin/poll [POST]
func (finHandler *FinHandler) Poll(g *gin.Context) {
	record := finHandler.currentFin(g)

	// A malformed body is tolerated (the whole request body is optional);
	// only a well-formed-but-invalid one is rejected.
	var request fin.PollRequest
	if g.Request.ContentLength > 0 {
		if err := g.ShouldBindJSON(&request); err != nil {
			log.Error(err)
			apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse poll request", "POST /fin/poll", err.Error())
			return
		}
	}

	if err := finHandler.repository.Touch(record.FinId, time.Now()); err != nil {
		// Polling is this protocol's liveness signal (§2.4) - a failure to
		// record it is logged, but must not block the Fin from actually
		// getting work.
		log.Warning("failed to update last-seen for fin ", record.FinId, ": ", err)
	}

	timeout := time.Duration(finHandler.config.LongPollTimeoutSeconds) * time.Second
	if finHandler.config.LongPollTimeoutSeconds <= 0 {
		timeout = 25 * time.Second
	}
	ctx, cancel := context.WithTimeout(g.Request.Context(), timeout)
	defer cancel()

	job, err := finHandler.queue.Claim(ctx, capabilityTypes(record.Capabilities), record.FinId)
	if err != nil {
		// Long-poll timed out (or the client disconnected) with no job
		// available - this is the expected, common case, not a failure.
		g.Status(http.StatusNoContent)
		return
	}

	g.JSON(http.StatusOK, fin.PollResponse{Job: job})
}

// SubmitResult
//
//	@Summary	submit the result of a claimed job
//	@Schemes
//	@Description	submit the result of a claimed job. Only the Fin the job is currently leased to may submit a result for it.
//	@Tags			fin
//	@Accept			json
//	@Produce		json
//	@Param			job_id	path	string				true	"job ID"
//	@Param			data	body	fin.ResultRequest	true	"job result"
//	@Success		204
//	@failure		400	{object}	api.Error
//	@failure		403	{object}	api.Error
//	@failure		404	{object}	api.Error
//	@Router			/fin/jobs/{job_id} [PUT]
func (finHandler *FinHandler) SubmitResult(g *gin.Context) {
	record := finHandler.currentFin(g)
	route := "PUT /fin/jobs/" + g.Param("job_id")

	jobId, err := uuid.Parse(g.Param("job_id"))
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse job ID", route, "")
		return
	}

	var request fin.ResultRequest
	if err := g.ShouldBindJSON(&request); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse job result", route, err.Error())
		return
	}
	if request.State != fin.JobStateSuccess && request.State != fin.JobStateFailure {
		apiError.SendErrorResponse(g, http.StatusBadRequest, "state must be \"success\" or \"failure\"", route, "")
		return
	}

	err = finHandler.queue.Submit(jobId, record.FinId, request.JobResult)
	if err != nil {
		finHandler.sendJobError(g, route, err)
		return
	}

	g.Status(http.StatusNoContent)
}

// StatusPing
//
//	@Summary	extend a claimed job's lease, and check for a pending cancellation instruction
//	@Schemes
//	@Description	extend a claimed job's lease (so a legitimately long-running job isn't requeued out from under the Fin still working on it), and check for a pending instruction such as cancellation
//	@Tags			fin
//	@Accept			json
//	@Produce		json
//	@Param			job_id	path		string					true	"job ID"
//	@Param			data	body		fin.StatusPingRequest	false	"progress"
//	@Success		200		{object}	fin.StatusPingResponse
//	@failure		400		{object}	api.Error
//	@failure		403		{object}	api.Error
//	@failure		404		{object}	api.Error
//	@Router			/fin/jobs/{job_id}/status [PATCH]
func (finHandler *FinHandler) StatusPing(g *gin.Context) {
	record := finHandler.currentFin(g)
	route := "PATCH /fin/jobs/" + g.Param("job_id") + "/status"

	jobId, err := uuid.Parse(g.Param("job_id"))
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse job ID", route, "")
		return
	}

	if g.Request.ContentLength > 0 {
		var request fin.StatusPingRequest
		if err := g.ShouldBindJSON(&request); err != nil {
			log.Error(err)
			apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse status ping", route, err.Error())
			return
		}
	}

	err = finHandler.queue.ExtendLease(jobId, record.FinId, finHandler.config.JobLeaseSeconds)
	if err != nil {
		finHandler.sendJobError(g, route, err)
		return
	}

	// Cancellation is not implemented yet (see
	// docs/adr/FIN-WEBHOOK-PROTOCOL-PROPOSAL.md §2.5) - Action is always
	// empty for now; this response shape exists so a Fin can start
	// checking it before SOARCA ever actually sets it.
	g.JSON(http.StatusOK, fin.StatusPingResponse{})
}

// Unregister
//
//	@Summary	delete this Fin's own registration
//	@Schemes
//	@Description	delete this Fin's own registration. The fin_id is inferred from the fin_token presented in the Authorization header - a Fin can only ever delete its own registration, so it never needs to name itself explicitly.
//	@Tags			fin
//	@Produce		json
//	@Success		204
//	@failure		404	{object}	api.Error
//	@Router			/fin/ [DELETE]
func (finHandler *FinHandler) Unregister(g *gin.Context) {
	record := finHandler.currentFin(g)
	route := "DELETE /fin/"

	if err := finHandler.repository.Unregister(record.FinId); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", route, "")
		return
	}

	g.Status(http.StatusNoContent)
}

// ############################################################################
// Read-only discovery (not Fin-authenticated: admin/dashboard reads)
// ############################################################################

// List
//
//	@Summary	list all currently-registered fins and their capabilities
//	@Schemes
//	@Description	list all currently-registered fins and their capabilities
//	@Tags			fin
//	@Produce		json
//	@Success		200	{object}	fin.ListResponse
//	@Router			/fin/ [GET]
func (finHandler *FinHandler) List(g *gin.Context) {
	records, err := finHandler.repository.List()
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError, "Failed to list fins", "GET /fin/", "")
		return
	}
	for i := range records {
		records[i].Stale = finHandler.isStale(records[i])
	}
	g.JSON(http.StatusOK, fin.ListResponse{Fins: records})
}

// Get
//
//	@Summary	look up a specific registered fin by id
//	@Schemes
//	@Description	look up a specific registered fin by id
//	@Tags			fin
//	@Produce		json
//	@Param			fin_id	path		string	true	"fin ID"
//	@Success		200		{object}	fin.Record
//	@failure		404		{object}	api.Error
//	@Router			/fin/{fin_id} [GET]
func (finHandler *FinHandler) Get(g *gin.Context) {
	finId := g.Param("fin_id")
	record, err := finHandler.repository.Get(finId)
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", "GET /fin/"+finId, "")
		return
	}
	record.Stale = finHandler.isStale(record)
	g.JSON(http.StatusOK, record)
}

// Delete
//
//	@Summary	forcibly remove a registered fin (admin)
//	@Schemes
//	@Description	forcibly remove a registered fin's record, e.g. one that is stale/offline and will never come back to unregister itself. This is an admin/dashboard action, not Fin-authenticated - unlike Unregister, it is not restricted to a fin removing its own registration.
//	@Tags			fin
//	@Produce		json
//	@Param			fin_id	path	string	true	"fin ID"
//	@Success		204
//	@failure		404	{object}	api.Error
//	@Router			/fin/{fin_id} [DELETE]
func (finHandler *FinHandler) Delete(g *gin.Context) {
	finId := g.Param("fin_id")
	route := "DELETE /fin/" + finId

	if err := finHandler.repository.Unregister(finId); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", route, "")
		return
	}

	g.Status(http.StatusNoContent)
}

// isStale reports whether record hasn't been seen (via /poll) within the
// configured staleness threshold - see Config.StaleAfterSeconds.
func (finHandler *FinHandler) isStale(record fin.Record) bool {
	staleAfter := defaultStaleAfter
	if finHandler.config.StaleAfterSeconds > 0 {
		staleAfter = time.Duration(finHandler.config.StaleAfterSeconds) * time.Second
	}
	return time.Since(record.LastSeen) > staleAfter
}

// ############################################################################
// Utility
// ############################################################################

// currentFin retrieves the fin.Record RequireFinToken resolved for this
// request. Only ever called from handlers registered behind that
// middleware, so the type assertion is always expected to succeed.
func (finHandler *FinHandler) currentFin(g *gin.Context) fin.Record {
	value, _ := g.Get(finContextKey)
	record, _ := value.(fin.Record)
	return record
}

func (finHandler *FinHandler) sendJobError(g *gin.Context, route string, err error) {
	log.Error(err)
	var notFound fin.ErrJobNotFound
	var notLeased fin.ErrJobNotLeasedToFin
	switch {
	case errors.As(err, &notFound):
		apiError.SendErrorResponse(g, http.StatusNotFound, "Job not found", route, "")
	case errors.As(err, &notLeased):
		apiError.SendErrorResponse(g, http.StatusForbidden, "Job is not leased to this fin", route, "")
	default:
		apiError.SendErrorResponse(g, http.StatusInternalServerError, "Failed to process job request", route, "")
	}
}

func bearerToken(g *gin.Context) (string, bool) {
	header := g.GetHeader("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	value := strings.TrimPrefix(header, prefix)
	if value == "" {
		return "", false
	}
	return value, true
}

func capabilityTypes(capabilities []fin.Capability) []string {
	types := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		types = append(types, capability.Type)
	}
	return types
}
