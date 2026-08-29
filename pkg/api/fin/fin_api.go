package fin

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	"soarca/internal/logger"
	"soarca/internal/storage"
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

const finContextKey = "fin_record"

type Config struct {
	RegistrationToken      string
	PollIntervalSeconds    int
	LongPollTimeoutSeconds int
	JobLeaseSeconds        int
	StaleAfter             time.Duration
}

// HandlerDependencies groups the dependencies needed to construct a FinHandler.
type HandlerDependencies struct {
	Store  storage.FinStore
	Queue  *queue.Queue
	Config Config
	GUID   guid.IGuid
}

type FinHandler struct {
	store  storage.FinStore
	queue  *queue.Queue
	config Config
	guid   guid.IGuid
}

func NewFinHandler(deps HandlerDependencies) *FinHandler {
	return &FinHandler{
		store:  deps.Store,
		queue:  deps.Queue,
		config: deps.Config,
		guid:   deps.GUID,
	}
}

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

	if err := finHandler.store.Create(g.Request.Context(), record); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError, "Failed to register fin", route, "")
		return
	}

	log.Info("registered fin ", record.FinId, " (", record.DisplayName, ") with capabilities ", capabilityTypes(record.Capabilities))
	g.JSON(http.StatusCreated, fin.RegisterResponse{FinId: record.FinId, FinToken: finToken, PollIntervalSeconds: finHandler.config.PollIntervalSeconds, LongPollTimeoutSeconds: finHandler.config.LongPollTimeoutSeconds, JobLeaseSeconds: finHandler.config.JobLeaseSeconds})
}

func (finHandler *FinHandler) RequireFinToken(g *gin.Context) {
	const route = "fin bearer auth"
	presentedToken, ok := bearerToken(g)
	if !ok {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "Missing or malformed Authorization header", route, "")
		g.Abort()
		return
	}

	record, err := finHandler.store.GetByTokenHash(g.Request.Context(), token.Hash(presentedToken))
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "Invalid or unknown fin token", route, "")
		g.Abort()
		return
	}

	g.Set(finContextKey, record)
	g.Next()
}

func (finHandler *FinHandler) Poll(g *gin.Context) {
	record := finHandler.currentFin(g)
	var request fin.PollRequest
	if g.Request.ContentLength > 0 {
		if err := g.ShouldBindJSON(&request); err != nil {
			log.Error(err)
			apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse poll request", "POST /fin/poll", err.Error())
			return
		}
	}
	if err := finHandler.store.Touch(g.Request.Context(), record.FinId, time.Now()); err != nil {
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
		g.Status(http.StatusNoContent)
		return
	}
	g.JSON(http.StatusOK, fin.PollResponse{Job: job})
}

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
	if err := finHandler.queue.Submit(jobId, record.FinId, request.JobResult); err != nil {
		finHandler.sendJobError(g, route, err)
		return
	}
	if err := finHandler.store.Touch(g.Request.Context(), record.FinId, time.Now()); err != nil {
		log.Warning("failed to update last-seen for fin ", record.FinId, ": ", err)
	}
	g.Status(http.StatusNoContent)
}

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
	if err := finHandler.queue.ExtendLease(jobId, record.FinId, finHandler.config.JobLeaseSeconds); err != nil {
		finHandler.sendJobError(g, route, err)
		return
	}
	if err := finHandler.store.Touch(g.Request.Context(), record.FinId, time.Now()); err != nil {
		log.Warning("failed to update last-seen for fin ", record.FinId, ": ", err)
	}
	g.JSON(http.StatusOK, fin.StatusPingResponse{})
}

func (finHandler *FinHandler) Unregister(g *gin.Context) {
	record := finHandler.currentFin(g)
	route := "DELETE /fin/"
	if err := finHandler.store.Delete(g.Request.Context(), record.FinId); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", route, "")
		return
	}
	g.Status(http.StatusNoContent)
}

func (finHandler *FinHandler) List(g *gin.Context) {
	records, err := finHandler.store.List(g.Request.Context())
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

func (finHandler *FinHandler) Get(g *gin.Context) {
	finId := g.Param("fin_id")
	record, err := finHandler.store.Get(g.Request.Context(), finId)
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", "GET /fin/"+finId, "")
		return
	}
	record.Stale = finHandler.isStale(record)
	g.JSON(http.StatusOK, record)
}

func (finHandler *FinHandler) Delete(g *gin.Context) {
	finId := g.Param("fin_id")
	route := "DELETE /fin/" + finId
	if err := finHandler.store.Delete(g.Request.Context(), finId); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", route, "")
		return
	}
	g.Status(http.StatusNoContent)
}

func (finHandler *FinHandler) isStale(record fin.Record) bool {
	return time.Since(record.LastSeen) > finHandler.config.StaleAfter
}

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
