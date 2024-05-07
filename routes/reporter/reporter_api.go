package reporter

import (
	"net/http"
	"soarca/internal/controller/informer"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
//	@Schemes		[]list
//	@Description	return all stored executions
//	@Tags			reporter
//	@Produce		json
//	@success		200	{array}	string
//	@error			400
//	@Router			/report/ [GET]
func (executionInformer *executionInformer) getExecutions(g *gin.Context) {
	executions := executionInformer.informer.GetExecutionsIds()
	g.JSON(http.StatusOK, executions)
}

// getExecutionReport GET handler for obtaining the information about an execution.
// Returns this to the gin context as a PlaybookExecutionReport object at soarca/model/api/reporter
//
//	@Summary		gets information about an ongoing playbook execution
//	@Schemes		soarca/models/api/PlaybookExecutionReport
//	@Description	return execution information
//	@Tags			reporter
//	@Produce		json
//	@success		200	PlaybookExecutionReport
//	@error			400
//	@Router			/report/:id [GET]
func (executionInformer *executionInformer) getExecutionReport(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to obtain execution for id: ", id)
	uuid := uuid.MustParse(id)

	executionEntry, err := executionInformer.informer.GetExecutionReport(uuid)
	if err != nil {
		log.Debug("Could not find execution for given id")
		SendErrorResponse(g, http.StatusBadRequest, "Could not find execution for given ID", "GET /report/{id}")
		return
	}

	executionEntryParsed, err := parseCachePlaybookEntry(executionEntry)
	if err != nil {
		log.Debug("Could not parse entry to reporter result model")
		SendErrorResponse(g, http.StatusInternalServerError, "Could not parse execution report", "GET /report/{id}")
		return
	}
	g.JSON(http.StatusOK, executionEntryParsed)
}

func SendErrorResponse(g *gin.Context, status int, message string, orginal_call string) {
	msg := gin.H{
		"status":        strconv.Itoa(status),
		"message":       message,
		"original-call": orginal_call,
	}
	g.JSON(status, msg)
}
