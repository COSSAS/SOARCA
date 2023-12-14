package application

import (
	"soarca/internal/capability"
	"soarca/internal/capability/ssh"
	"soarca/internal/decomposer"
	"soarca/internal/executer"
	"soarca/internal/guid"
	"soarca/routes"

	"github.com/gin-gonic/gin"
)

func InitialiseCore(app *gin.Engine) error {
	ssh := new(ssh.SshCapability)
	capabilities := map[string]capability.ICapability{ssh.GetType(): ssh}
	executer := executer.New(capabilities)
	guid := new(guid.Guid)
	decompose := decomposer.New(executer, guid)

	err := routes.Api(app, decompose)
	if err != nil {
		log.Error(err)
		return err
	}
	routes.Logging(app)
	routes.Swagger(app)
	return err
}
