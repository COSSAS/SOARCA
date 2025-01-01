package caldera

import (
	"soarca/pkg/core/capability/caldera/api/client"
	"sync"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"soarca/pkg/utils"
)

type calderaInstance struct {
	send client.Caldera
}

var cInstance *calderaInstance
var instanceLock = &sync.Mutex{}

func newCalderaInstance() (*calderaInstance, error) {
	var config = client.DefaultTransportConfig()
	config.Host = utils.GetEnv("CALDERA_HOST", "") + ":" + utils.GetEnv("CALDERA_PORT", "")
	config.Schemes = []string{"http"}
	var calderaClient = client.NewHTTPClientWithConfig(nil, config)
	return &calderaInstance{*calderaClient}, nil
}

func GetCalderaInstance() (*calderaInstance, error) {
	if cInstance == nil {
		instanceLock.Lock()
		defer instanceLock.Unlock()
		if cInstance == nil {
			var instance, err = newCalderaInstance()
			if err != nil {
				return nil, err
			}
			cInstance = instance
		}
	}
	return cInstance, nil
}

func authenticateCaldera(operation *runtime.ClientOperation) {
	var apiKey = utils.GetEnv("CALDERA_API_KEY", "")
	operation.AuthInfo = httptransport.APIKeyAuth("KEY", "header", apiKey)
}
