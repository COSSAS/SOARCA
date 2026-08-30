package playbook

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/orchestrator"
	"soarca/internal/store"
	"soarca/pkg/cacao"
	"strconv"

	"github.com/gin-gonic/gin"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type playbookHandler struct {
	playbooks orchestrator.PlaybookService
}

func NewPlaybookHandler(playbooks orchestrator.PlaybookService) *playbookHandler {
	return &playbookHandler{playbooks: playbooks}
}

func (handler *playbookHandler) GetAllPlaybooks(g *gin.Context) {
	log.Trace("Trying to obtain all playbook IDs")

	returnListIDs, err := handler.playbooks.ListPlaybooks(g.Request.Context())
	if err != nil {
		log.Debug("Could not obtain any Playbooks", err)
		SendErrorResponse(g, http.StatusBadRequest, "Could not obtain any IDs", "GET /playbook")
		return
	}

	g.JSON(http.StatusOK, returnListIDs)
}

func (handler *playbookHandler) GetAllPlaybookMetas(g *gin.Context) {
	log.Trace("Trying to obtain all playbook IDs")

	returnListIDs, err := handler.playbooks.ListPlaybookMetas(g.Request.Context())
	if err != nil {
		log.Debug("Could not obtain any PlaybookMetas", err)
		SendErrorResponse(g, http.StatusBadRequest, "Could not obtain any IDs", "GET /playbook/meta")
		return
	}

	g.JSON(http.StatusOK, returnListIDs)
}

func (handler *playbookHandler) SubmitPlaybook(g *gin.Context) {
	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Trace("Submit playbook Endpoint has failed: ", err.Error())
		SendErrorResponse(g, http.StatusBadRequest, "Failed to marshall json on server side", "POST /playbook")
		return
	}
	var playbook cacao.Playbook
	if err := json.Unmarshal(jsonData, &playbook); err != nil {
		SendErrorResponse(g, http.StatusBadRequest, "Could not create playbook. Is the playbook correct?", "POST /playbook")
		return
	}
	if err := handler.playbooks.CreatePlaybook(g.Request.Context(), &playbook); err != nil {
		if err == storage.ErrConflict {
			SendErrorResponse(g, http.StatusConflict, "Provided duplicate playbook, already in database", "POST /playbook")
			return
		}
		SendErrorResponse(g, http.StatusBadRequest, "Could not create playbook. Is the playbook correct?", "POST /playbook")
		return
	}
	g.JSON(http.StatusCreated, playbook)
}

func (handler *playbookHandler) GetPlaybookByID(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to obtain playbook for id: ", id)

	playbook, err := handler.playbooks.GetPlaybook(g.Request.Context(), id)
	if err != nil {
		log.Debug("Could not find document for given id")
		SendErrorResponse(g, http.StatusNotFound, "Could not find playbook for given ID", "GET /playbook/{id}")
		return
	}
	g.JSON(http.StatusOK, playbook)
}

func (handler *playbookHandler) UpdatePlaybookByID(g *gin.Context) {
	id := g.Param("id")
	log.Trace("Trying to update playbook for id: ", id)

	jsonData, err := io.ReadAll(g.Request.Body)
	if err != nil {
		log.Debug("Update playbook Endpoint has failed: ", err.Error())
		SendErrorResponse(g, http.StatusBadRequest, "Failed to marshall json on server sider", "PUT /playbook/{id}")
		return
	}
	var updatedPlaybook cacao.Playbook
	if err := json.Unmarshal(jsonData, &updatedPlaybook); err != nil {
		SendErrorResponse(g, http.StatusBadRequest, "Could not find playbook for given ID", "PUT /playbook/{id}")
		return
	}
	if err := handler.playbooks.UpdatePlaybook(g.Request.Context(), id, &updatedPlaybook); err != nil {
		if err == storage.ErrNotFound {
			SendErrorResponse(g, http.StatusNotFound, "Could not find playbook for given ID", "PUT /playbook/{id}")
			return
		}
		SendErrorResponse(g, http.StatusBadRequest, "Could not find playbook for given ID", "PUT /playbook/{id}")
		return
	}
	g.JSON(http.StatusOK, updatedPlaybook)
}

func (handler *playbookHandler) DeleteByPlaybookID(g *gin.Context) {
	id := g.Param("id")
	err := handler.playbooks.DeletePlaybook(g.Request.Context(), id)
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
