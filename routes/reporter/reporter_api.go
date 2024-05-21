package reporter

import (
	"net/http"
	"soarca/internal/controller/informer"

	"reflect"
	"soarca/routes/error"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"soarca/logger"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// A PlaybookController implements the playbook API endpoints is dependent on a database.
type executionInformer struct {
	informer informer.IExecutionInformer
}

// NewPlaybookController makes a new instance of playbookControler
func NewExecutionInformer(informer informer.IExecutionInformer) *executionInformer {
	return &executionInformer{informer: informer}
}

// getExecutions GET handler for obtaining all the executions that can be retrieved.
// Returns this to the gin context as a list if execution IDs in json format
//
//	@Summary	gets all the UUIDs for the executions that can be retireved
//	@Schemes
//	@Description	return all stored executions
//	@Tags			reporter
//	@Produce		json
//	@success		200	{array}		api.PlaybookExecutionReport
//	@failure		400	{object}	api.Error
//	@Router			/reporter [GET]
func (executionInformer *executionInformer) getExecutions(g *gin.Context) {
	executions, err := executionInformer.informer.GetExecutions()
	if err != nil {
		log.Debug("Could not get executions from informer")
		error.SendErrorResponse(g, http.StatusInternalServerError, "Could not get executions from informer", "GET /report/", "")
		return
	}
	g.JSON(http.StatusOK, executions)
}

// getExecutionReport GET handler for obtaining the information about an execution.
// Returns this to the gin context as a PlaybookExecutionReport object at soarca/model/api/reporter
//
//	@Summary	gets information about an ongoing playbook execution
//	@Schemes
//	@Description	return execution information
//	@Tags			reporter
//	@Produce		json
//	@Param			id	path		string	true	"execution identifier"
//	@success		200	{object}	api.PlaybookExecutionReport
//	@failure		400	{object}	api.Error
//	@Router			/reporter/{id} [GET]
func (executionInformer *executionInformer) getExecutionReport(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to obtain execution for id: ", id)
	uuid, err := uuid.Parse(id)
	if err != nil {
		log.Debug("Could not parse id parameter for request")
		error.SendErrorResponse(g, http.StatusBadRequest, "Could not parse id parameter for request", "GET /report/{id}", "")
		return
	}

	executionEntry, err := executionInformer.informer.GetExecutionReport(uuid)
	if err != nil {
		log.Debug("Could not find execution for given id")
		error.SendErrorResponse(g, http.StatusBadRequest, "Could not find execution for given ID", "GET /report/{id}", "")
		return
	}

	executionEntryParsed, err := parseCachePlaybookEntry(executionEntry)
	if err != nil {
		log.Debug("Could not parse entry to reporter result model")
		error.SendErrorResponse(g, http.StatusInternalServerError, "Could not parse execution report", "GET /report/{id}", "")
		return
	}
	g.JSON(http.StatusOK, executionEntryParsed)
}
