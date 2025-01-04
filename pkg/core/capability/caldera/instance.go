// This file contains the bundled methods and structs to handle the Caldera server instance and the singleton method wrapped around it.
package caldera

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"soarca/pkg/core/capability/caldera/api/client"
	"soarca/pkg/utils"
	"sync"
)

// calderaInstance is a singleton struct that models the Caldera server Instance.
//
// It contains a field that serves as entrypoint to the auto generated Caldera OpenAPI client.
type calderaInstance struct {
	// send is the entrypoint to the auto generated Caldera OpenAPI client.
	send client.Caldera
}

// cInstance is the calderaInstance singleton instance.
var cInstance *calderaInstance

// instanceLock is used as the Mutex lock for the calderaInstance sigleton.
var instanceLock = &sync.Mutex{}

// newCalderaInstance builds and configures a new calderaInstance.
//
// It retrieves the Caldera server instance host URL and port from the environment variables with names 'CALDERA_HOST' and 'CALDERA_PORT'.
//
// Plans are to check if the server instance is up and healty.
func newCalderaInstance() (*calderaInstance, error) {
	var config = client.DefaultTransportConfig()
	config.Host = utils.GetEnv("CALDERA_HOST", "") + ":" + utils.GetEnv("CALDERA_PORT", "")
	config.Schemes = []string{"http"}
	var calderaClient = client.NewHTTPClientWithConfig(nil, config)
	return &calderaInstance{*calderaClient}, nil
}

// GetCalderaInstance is the global access method to retrieve the calderaInstance singleton, and should be the only method used everywhere else.
//
// If the singleton is not instanciated, it will create it first.
// It will only return an error if the creation of the singleton fails, else it will return the singleton instance.
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

// authenticateCaldera is used in the connection requests and adds the authentication info to the request.
//
// The authentication info now consists of an API key retrieved from the environment variables, namely 'CALDERA_API_KEY'.
func authenticateCaldera(operation *runtime.ClientOperation) {
	var apiKey = utils.GetEnv("CALDERA_API_KEY", "")
	operation.AuthInfo = httptransport.APIKeyAuth("KEY", "header", apiKey)
}
