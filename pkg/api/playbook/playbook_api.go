package playbook

import (
	"io"
	"net/http"
	"reflect"
	"soarca/internal/controller/database"
	"soarca/internal/logger"
	"strconv"

	playbookrepository "soarca/internal/database/playbook"

	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// a playbookHandler implements the playbook api endpoints is dependent on a database.
type playbookHandler struct {
	playbookRepo playbookrepository.IPlaybookRepository
}

// NewPlaybookHandler makes a new instance of NewPlaybookHandler
func NewPlaybookHandler(controller database.IController) *playbookHandler {
	return &playbookHandler{playbookRepo: controller.GetDatabaseInstance()}
}

// GetAllPlaybooks GET handler for obtaining all the playbooks in the database and return this to the gin context in json format
//
//	@Summary	gets all the UUIDs for the stored playbooks
//	@Schemes
//	@Description	return all stored playbooks default limit:100
//	@Tags			playbook
//	@Produce		json
//	@success		200	{array}		cacao.Playbook
//	@failure		400	{object}	api.Error
//	@Router			/playbook/ [GET]
func (handler *playbookHandler) GetAllPlaybooks(g *gin.Context) {
	log.Trace("Trying to obtain all playbook IDs")

	returnListIDs, err := handler.playbookRepo.GetPlaybooks()
	if err != nil {
		log.Debug("Could not obtain any PlaybookMetas", err)
		SendErrorResponse(g, http.StatusBadRequest, "Could not obtain any IDs", "GET /playbook/meta")
		return
	}

	g.JSON(http.StatusOK, returnListIDs)
}

// GetAllPlaybookMetas GET handler for obtaining all the meta data of all the stored playbooks
// in the database and return this to the gin context in json format
//
//	@Summary	gets all the meta information for the stored playbooks
//	@Schemes
//	@Description	get playbook meta information for playbook
//	@Tags			playbook
//	@Produce		json
//	@success		200	{array}		api.PlaybookMeta
//	@failure		400	{object}	api.Error
//	@Router			/playbook/meta [GET]
func (handler *playbookHandler) GetAllPlaybookMetas(g *gin.Context) {
	log.Trace("Trying to obtain all playbook IDs")

	returnListIDs, err := handler.playbookRepo.GetPlaybookMetas()
	if err != nil {
		log.Debug("Could not obtain any PlaybookMetas", err)
		SendErrorResponse(g, http.StatusBadRequest, "Could not obtain any IDs", "GET /playbook/meta")
		return
	}

	g.JSON(http.StatusOK, returnListIDs)
}

// SubmitPlaybook POST handler for creating playbooks.
//
//	@Summary	submit playbook via the api
//	@Schemes
//	@Description	submit a new playbook api
//	@Tags			playbook
//	@Produce		json
//	@Accept			json
//	@Param			data	body		cacao.Playbook	true	"playbook"
//	@Success		200		{object}	cacao.Playbook
//	@failure		400		{object}	api.Error
//	@Router			/playbook/ [POST]
func (handler *playbookHandler) SubmitPlaybook(g *gin.Context) {
	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Trace("Submit playbook Endpoint has failed: ", err.Error())
		SendErrorResponse(g, http.StatusBadRequest, "Failed to marshall json on server side", "POST /playbook")
		return
	}
	playbook, err := handler.playbookRepo.Create(&jsonData)
	if err != nil {
		log.Debug("Submit playbook Endpoint has failed:", err.Error())
		if err.Error() == "duplicate" {
			SendErrorResponse(g, http.StatusConflict, "Provided duplicate playbook, already in database", "POST /playbook")
		} else {
			SendErrorResponse(g, http.StatusBadRequest, "Could not create playbook. Is the playbook correct?", "POST /playbook")
		}
		return
	}
	g.JSON(http.StatusCreated, playbook)
}

// GetPlaybookByID GET handler that finds playbook by id
//
//	@Summary	get CACAO playbook by its ID
//	@Schemes
//	@Description	get playbook by ID
//	@Tags			playbook
//	@Produce		json
//	@Accept			json
//	@Param			id	path		string	true	"playbook ID"
//	@Success		200	{object}	cacao.Playbook
//	@failure		400	{object}	api.Error
//	@Router			/playbook/{id} [GET]
func (handler *playbookHandler) GetPlaybookByID(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to obtain playbook for id: ", id)

	playbook, err := handler.playbookRepo.Read(id)
	if err != nil {
		log.Debug("Could not find document for given id")
		SendErrorResponse(g, http.StatusBadRequest, "Could not find playbook for given ID", "GET /playbook/{id}")
		return
	}
	g.JSON(http.StatusOK, playbook)
}

// UpdatePlaybookByID PUT handler that allows updating playbook object by ID.
//
//	@Summary	update playbook
//	@Schemes
//	@Description	update playbook by Id
//	@Tags			playbook
//	@Produce		json
//	@Accept			json
//	@Param			id		path		string			true	"playbook Id"
//	@Param			data	body		cacao.Playbook	true	"playbook"
//	@Success		200		{object}	cacao.Playbook
//	@failure		400		{object}	api.Error
//	@Router			/playbook/{id} [PUT]
func (handler *playbookHandler) UpdatePlaybookByID(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to update playbook for id: ", id)

	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Debug("Update playbook Endpoint has failed: ", err.Error())
		SendErrorResponse(g, http.StatusBadRequest, "Failed to marshall json on server sider", "PUT /playbook/{id}")
		return
	}
	updatedData, err := handler.playbookRepo.Update(id, &jsonData)
	if err != nil {
		log.Trace("Could not find document for given ")
		SendErrorResponse(g, http.StatusBadRequest, "Could not find playbook for given ID", "PUT /playbook/{id}")
		return
	}
	g.JSON(http.StatusOK, updatedData)
}

// DeleteByPlaybookID DELETE handler for deleting playbook by ID.
//
//	@Summary	delete playbook by Id
//	@Schemes
//	@Description	delete playbook by Id
//	@Tags			playbook
//	@Produce		json
//	@Accept			json
//	@Param			id	path	string	true	"playbook ID"
//	@Success		200
//	@failure		400	{object}	api.Error
//	@Router			/playbook/{id} [DELETE]
func (handler *playbookHandler) DeleteByPlaybookID(g *gin.Context) {
	id := g.Param("id")
	err := handler.playbookRepo.Delete(id)
	if err != nil {
		log.Debug("Something when wrong tying to delete the playbook object. Does the object exists?")
		SendErrorResponse(g, http.StatusBadRequest, "Could not delete object", "DELETE /playbook/{id}")
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
