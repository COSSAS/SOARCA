package reporter

import (
	"net/http"
	"reflect"
	"soarca/internal/logger"
	execsvc "soarca/internal/runs"
	"soarca/internal/transport/http/handlers/error"
	api "soarca/internal/transport/http/schema"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// reportHandler implements the handler functions that can be called by the gin api is dependent on a database.
type reportHandler struct {
	runs execsvc.Runner
}

// NewReportHandler makes a new instance of playbookControler
func NewReportHandler(runner execsvc.Runner) *reportHandler {
	return &reportHandler{runs: runner}
}

// GetRuns GET handler for obtaining all the runs that can be retrieved.
// Returns this to the gin context as a list if run IDs in json format
//
//	@Summary	gets all the UUIDs for the runs that can be retireved
//	@Schemes
//	@Description	return all stored runs
//	@Tags			reporter
//	@Produce		json
//	@success		200	{array}		api.PlaybookRunReport
//	@failure		400	{object}	api.Error
//	@Router			/reporter [GET]
func (reportHandler *reportHandler) GetRuns(g *gin.Context) {
	runs, err := reportHandler.runs.List(g.Request.Context())
	if err != nil {
		log.Debug("Could not get runs from run state")
		error.SendErrorResponse(g, http.StatusInternalServerError, "Could not get runs", "GET /reporter/", "")
		return
	}

	runsParsed := []api.PlaybookRunReport{}
	for _, runEntry := range runs {
		runEntryParsed, err := parseRunStateEntry(runEntry)
		if err != nil {
			log.Debug("Could not parse entry to reporter result model")
			log.Error(err)
			error.SendErrorResponse(g, http.StatusInternalServerError, "Could not parse run report", "GET /reporter/", "")
			return
		}
		runsParsed = append(runsParsed, runEntryParsed)
	}

	g.JSON(http.StatusOK, runsParsed)
}

// GetRunReport GET handler for obtaining the information about an run.
// Returns this to the gin context as a PlaybookRunReport object at soarca/model/api/reporter
//
//	@Summary	gets information about an ongoing playbook run
//	@Schemes
//	@Description	return run information
//	@Tags			reporter
//	@Produce		json
//	@Param			id	path		string	true	"run identifier"
//	@success		200	{object}	api.PlaybookRunReport
//	@failure		400	{object}	api.Error
//	@Router			/reporter/{id} [GET]
func (handler *reportHandler) GetRunReport(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to obtain run for id: ", id)
	uuid, err := uuid.Parse(id)
	if err != nil {
		log.Debug("Could not parse id parameter for request")
		error.SendErrorResponse(g, http.StatusBadRequest, "Could not parse id parameter for request", "GET /reporter/"+id, err.Error())
		return
	}

	runEntry, err := handler.runs.Report(g.Request.Context(), uuid)
	if err != nil {
		log.Debug("Could not find run for given id")
		log.Error(err)
		error.SendErrorResponse(g, http.StatusBadRequest, "Could not find run for given ID", "GET /reporter/"+id, "")
		return
	}

	runEntryParsed, err := parseRunStateEntry(runEntry)
	if err != nil {
		log.Debug("Could not parse entry to reporter result model")
		log.Error(err)
		error.SendErrorResponse(g, http.StatusInternalServerError, "Could not parse run report", "GET /reporter/"+id, "")
		return
	}
	g.JSON(http.StatusOK, runEntryParsed)
}
