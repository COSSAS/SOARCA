package workflow

import (
	"io"
	"net/http"
	"strconv"

	workflowRepository "soarca/database/workflow"

	"github.com/gin-gonic/gin"
)

// A WorkflowController implements the workflow API endpoints is dependent on a database.
type workflowController struct {
	workflowRepo workflowRepository.IWorkflowRepository
}

// NewWorkflowController makes a new instance of workflowControler
func NewWorkflowController(workflowRepo workflowRepository.IWorkflowRepository) *workflowController {
	return &workflowController{workflowRepo: workflowRepo}
}

// getAllsWorkflows GET handler for obtaining all the workflows in the database and return this to the gin context in json format
// @Summary gets all the UUIDs for the stored workflows
// @Schemes
// @Description get playbook meta information for workflow
// @Tags workflow
// @Produce json
// @success 200 {array} cacao.Playbook
// @error	400
// @Router /workflow/ [GET]
func (workflowctrl *workflowController) getAllWorkflows(g *gin.Context) {
	log.Trace("Trying to obtain all workflow IDs")

	returnListIDs, err := workflowctrl.workflowRepo.GetWorkflows()
	if err != nil {
		log.Debug("Could not obtain any PlaybookMetas", err)
		SendErrorResponse(g, http.StatusBadRequest, "Could not obtain any IDs", "GET /workflow/meta")
		return
	}

	g.JSON(http.StatusOK, returnListIDs)
}

// getAllsWorkflows GET handler for obtaining all the meta data of all the stored workflows
// in the database and return this to the gin context in json format
// @Summary gets all the c for the stored workflows
// @Schemes
// @Description get playbook meta information for workflow
// @Tags workflow
// @Produce json
// @success 200 {array} api.PlaybookMeta
// @error	400
// @Router /workflow/meta [GET]
func (workflowctrl *workflowController) getAllWorkFlowMetas(g *gin.Context) {
	log.Trace("Trying to obtain all workflow IDs")

	returnListIDs, err := workflowctrl.workflowRepo.GetWorkflowMetas()
	if err != nil {
		log.Debug("Could not obtain any PlaybookMetas", err)
		SendErrorResponse(g, http.StatusBadRequest, "Could not obtain any IDs", "GET /workflow/meta")
		return
	}

	g.JSON(http.StatusOK, returnListIDs)
}

// submitWorkflow POST handler for creating worflows.
// @Summary submit workflow via the api
// @Schemes
// @Description submit a new workflow api
// @Tags workflow
// @Produce json
// @Accept json
// @Param data body cacao.Playbook true "workflow"
// @Success 200  {object} cacao.Playbook
// @error 400
// @Router /workflow/ [POST]
func (workflowctrl *workflowController) submitWorkflow(g *gin.Context) {
	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Trace("Submit Workflow Endpoint has failed: ", err.Error())
		SendErrorResponse(g, http.StatusBadRequest, "Failed to marshall json on server side", "POST /workflow")
		return
	}
	playbook, err := workflowctrl.workflowRepo.Create(&jsonData)
	if err != nil {
		log.Debug("Submit Workflow Endpoint has failed:", err.Error())
		if err.Error() == "duplicate" {
			SendErrorResponse(g, http.StatusConflict, "Provided duplicate workflow, already in database", "POST /workflow")
		} else {
			SendErrorResponse(g, http.StatusInternalServerError, "Internal server error, could not create Workflow. Is the playbook correct?", "POST /workflow")
		}
		return
	}
	g.JSON(http.StatusCreated, playbook)
}

// getWorkflowbyID GET handler that finds workflow by id
// @Summary get CACAO playbook workflow by its ID
// @Schemes
// @Description get workflow by ID
// @Tags workflow
// @Produce json
// @Accept json
// @Param id path string true "workflow ID"
// @Success 200  {object} cacao.Playbook
// @error 400
// @Router /workflow/{id} [GET]
func (workflowctrl *workflowController) getWorkflowByID(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to obtain Workflow for id: ", id)

	return_workflow, err := workflowctrl.workflowRepo.Read(id)
	if err != nil {
		log.Debug("Could not find document for given id")
		SendErrorResponse(g, http.StatusBadRequest, "Could not find workflow for given ID", "GET /workflow/{id}")
		return
	}
	g.JSON(http.StatusOK, return_workflow)
}

// updateWorkFlowbyID PUT handler that allows updating workflow object by ID.
// @Summary update workflow
// @Schemes
// @Description update workflow by ID
// @Tags workflow
// @Produce json
// @Accept json
// @Param id path string true "workflow ID"
// @Param data body cacao.Playbook true "workflow"
// @Success 200  {object} cacao.Playbook
// @error 400
// @Router /workflow/{id} [PUT]
func (workflowctrl *workflowController) updateWorkflowByID(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to update Workflow for id: ", id)

	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Debug("Update Workflow Endpoint has failed: ", err.Error())
		SendErrorResponse(g, http.StatusBadRequest, "Failed to marshall json on server sider", "PUT /workflow/{id}")
		return
	}
	updated_data, err := workflowctrl.workflowRepo.Update(id, &jsonData)
	if err != nil {
		log.Trace("Could not find document for given ")
		SendErrorResponse(g, http.StatusBadRequest, "Could not find workflow for given ID", "PUT /workflow/{id}")
		return
	}
	g.JSON(http.StatusOK, updated_data)
}

// deleteWorkflowbyID DELETE handler for deleting workflow by ID.
// @Summary delete worflow
// @Schemes
// @Description delete workflow by ID
// @Tags workflow
// @Produce json
// @Accept json
// @Param id path string true "workflow ID"
// @Success 200
// @error 400
// @Router /workflow/{id} [DELETE]
func (workflowctrl *workflowController) deleteWorkflowByID(g *gin.Context) {
	id := g.Param("id")
	err := workflowctrl.workflowRepo.Delete(id)
	if err != nil {
		log.Debug("Something when wrong tying to delete the workflow object. Does the object exists?")
		SendErrorResponse(g, http.StatusBadRequest, "Could not delete object", "DELETE /workflow/{id}")
		return
	}
	g.Status(http.StatusOK)
}

func SendErrorResponse(g *gin.Context, status int, message string, orginal_call string) {
	msg := gin.H{
		"status":        strconv.Itoa(status),
		"message":       message,
		"original-call": orginal_call,
	}
	g.JSON(status, msg)
}
