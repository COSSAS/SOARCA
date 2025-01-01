package caldera 

import (
	"os"
	"soarca/pkg/core/capability/caldera/api/client"
	"sync"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime"
)

type calderaInstance struct {
	send client.Caldera
}

var cInstance *calderaInstance
var instanceLock = &sync.Mutex{}

func newCalderaInstance() (*calderaInstance, error) {
	scheme, _ := os.LookupEnv("CALDERA_SCHEME")
	var config = client.DefaultTransportConfig()
	config.Host, _ = os.LookupEnv("CALDERA_HOST")
	config.Schemes = []string{scheme}
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
	operation.AuthInfo = httptransport.APIKeyAuth("KEY", "header", "ADMIN123")
}
