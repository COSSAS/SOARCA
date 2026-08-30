package fin

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	"soarca/internal/logger"
	"soarca/internal/orchestrator"
	apiError "soarca/internal/transport/http/handlers/error"
	"soarca/pkg/fins/protocol"

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

// FinHandler is the HTTP adapter for FIN operations.
type FinHandler struct {
	registry    orchestrator.FinRegistry
	workService orchestrator.FinWorkService
	config      Config
}

// NewFinHandler creates a new FIN HTTP handler with service dependencies.
func NewFinHandler(registry orchestrator.FinRegistry, workService orchestrator.FinWorkService, config Config) *FinHandler {
	return &FinHandler{
		registry:    registry,
		workService: workService,
		config:      config,
	}
}

func (h *FinHandler) Register(g *gin.Context) {
	const route = "POST /fin/register"

	var request fin.RegisterRequest
	if err := g.ShouldBindJSON(&request); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse registration request", route, err.Error())
		return
	}

	finID, finToken, err := h.registry.RegisterFin(g.Request.Context(), request)
	if err != nil {
		h.sendRegistrationError(g, route, err)
		return
	}

	g.JSON(http.StatusCreated, fin.RegisterResponse{
		FinId:                    finID,
		FinToken:                 finToken,
		PollIntervalSeconds:      h.config.PollIntervalSeconds,
		LongPollTimeoutSeconds:   h.config.LongPollTimeoutSeconds,
		JobLeaseSeconds:          h.config.JobLeaseSeconds,
	})
}

func (h *FinHandler) RequireFinToken(g *gin.Context) {
	const route = "fin bearer auth"
	presentedToken, ok := bearerToken(g)
	if !ok {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "Missing or malformed Authorization header", route, "")
		g.Abort()
		return
	}

	// Validate that the token is registered
	if _, err := h.registry.ValidateToken(g.Request.Context(), presentedToken); err != nil {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "Invalid or unknown fin token", route, "")
		g.Abort()
		return
	}

	// Store the token in context for handlers to use
	g.Set(finContextKey, presentedToken)
	g.Next()
}

func (h *FinHandler) Poll(g *gin.Context) {
	finToken, ok := h.getFinToken(g)
	if !ok {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "FIN token not found in context", "POST /fin/poll", "")
		return
	}

	var request fin.PollRequest
	if g.Request.ContentLength > 0 {
		if err := g.ShouldBindJSON(&request); err != nil {
			log.Error(err)
			apiError.SendErrorResponse(g, http.StatusBadRequest, "Failed to parse poll request", "POST /fin/poll", err.Error())
			return
		}
	}

	job, err := h.workService.PollJob(g.Request.Context(), finToken, request)
	if err != nil {
		// Poll timeout or context cancellation -> return no content
		g.Status(http.StatusNoContent)
		return
	}

	g.JSON(http.StatusOK, fin.PollResponse{Job: *job})
}

func (h *FinHandler) SubmitResult(g *gin.Context) {
	finToken, ok := h.getFinToken(g)
	if !ok {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "FIN token not found in context", "PUT /fin/jobs/:job_id", "")
		return
	}

	route := "PUT /fin/jobs/" + g.Param("job_id")
	jobID, err := uuid.Parse(g.Param("job_id"))
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

	if err := h.workService.SubmitJobResult(g.Request.Context(), finToken, jobID, request.JobResult); err != nil {
		h.sendJobError(g, route, err)
		return
	}

	g.Status(http.StatusNoContent)
}

func (h *FinHandler) StatusPing(g *gin.Context) {
	finToken, ok := h.getFinToken(g)
	if !ok {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "FIN token not found in context", "PATCH /fin/jobs/:job_id/status", "")
		return
	}

	route := "PATCH /fin/jobs/" + g.Param("job_id") + "/status"
	jobID, err := uuid.Parse(g.Param("job_id"))
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

	if err := h.workService.HeartbeatJob(g.Request.Context(), finToken, jobID); err != nil {
		h.sendJobError(g, route, err)
		return
	}

	g.JSON(http.StatusOK, fin.StatusPingResponse{})
}

func (h *FinHandler) Unregister(g *gin.Context) {
	finToken, ok := h.getFinToken(g)
	if !ok {
		apiError.SendErrorResponse(g, http.StatusUnauthorized, "FIN token not found in context", "DELETE /fin/", "")
		return
	}

	route := "DELETE /fin/"
	if err := h.registry.UnregisterFin(g.Request.Context(), finToken); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", route, "")
		return
	}

	g.Status(http.StatusNoContent)
}

func (h *FinHandler) List(g *gin.Context) {
	records, err := h.registry.ListFins(g.Request.Context())
	if err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusInternalServerError, "Failed to list fins", "GET /fin/", "")
		return
	}
	g.JSON(http.StatusOK, fin.ListResponse{Fins: records})
}

func (h *FinHandler) Get(g *gin.Context) {
	finID := g.Param("fin_id")
	record, err := h.registry.GetFin(g.Request.Context(), finID)
	if err != nil {
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", "GET /fin/"+finID, "")
		return
	}
	g.JSON(http.StatusOK, record)
}

func (h *FinHandler) Delete(g *gin.Context) {
	finID := g.Param("fin_id")
	route := "DELETE /fin/" + finID
	if err := h.registry.DeleteFin(g.Request.Context(), finID); err != nil {
		log.Error(err)
		apiError.SendErrorResponse(g, http.StatusNotFound, "Fin not found", route, "")
		return
	}
	g.Status(http.StatusNoContent)
}

// ============================================================================
// Helper methods
// ============================================================================

func (h *FinHandler) getFinToken(g *gin.Context) (string, bool) {
	value, ok := g.Get(finContextKey)
	if !ok {
		return "", false
	}
	token, ok := value.(string)
	return token, ok
}

func (h *FinHandler) sendRegistrationError(g *gin.Context, route string, err error) {
	log.Warning(err)

	var errRegistrationDisabled fin.ErrRegistrationDisabled
	var errTokenInvalid fin.ErrRegistrationTokenInvalid
	var errNoCapabilities fin.ErrNoCapabilities
	var errCapabilityTypeEmpty fin.ErrCapabilityTypeEmpty

	switch {
	case errors.As(err, &errRegistrationDisabled):
		apiError.SendErrorResponse(g, http.StatusServiceUnavailable, "Fin registration is not configured", route, "")
	case errors.As(err, &errTokenInvalid):
		apiError.SendErrorResponse(g, http.StatusForbidden, err.Error(), route, "")
	case errors.As(err, &errNoCapabilities):
		apiError.SendErrorResponse(g, http.StatusBadRequest, "At least one capability is required", route, "")
	case errors.As(err, &errCapabilityTypeEmpty):
		apiError.SendErrorResponse(g, http.StatusBadRequest, "Every capability requires a non-empty type", route, "")
	default:
		apiError.SendErrorResponse(g, http.StatusInternalServerError, "Failed to register fin", route, "")
	}
}

func (h *FinHandler) sendJobError(g *gin.Context, route string, err error) {
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
