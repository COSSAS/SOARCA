package jupyter

import (
	"encoding/base64"
	"os"
	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/jupyter"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelloWorld(t *testing.T) {
	notebook_file, err := os.ReadFile("helloworld.ipynb")
	assert.Nil(t, err)
	notebook_b64 := base64.StdEncoding.EncodeToString(notebook_file)
	addresses := cacao.Addresses{"url": []string{"http://localhost:7000"}}
	command := cacao.Command{
		Type:       cacao.CommandTypeJupyter,
		CommandB64: notebook_b64,
	}
	jupyter_capability := jupyter.JupyterCapability{}
	context := capability.Context{Command: command, Target: cacao.AgentTarget{Address: addresses}}
	metadata := execution.Metadata{}
	variables, err := jupyter_capability.Execute(metadata, context)
	assert.Nil(t, err)

	assert.Len(t, variables, 1)
	variable, ok := variables["__hello__"]
	assert.True(t, ok)
	assert.NotNil(t, variable)
	assert.Equal(t, variable.Name, "__hello__")
	assert.Equal(t, variable.Value, "world")
}
func TestCorrupted(t *testing.T) {
	notebook_file, err := os.ReadFile("helloworld.ipynb")
	assert.Nil(t, err)
	notebook_file[20] ^= 8
	notebook_file[10] ^= 8
	notebook_b64 := base64.StdEncoding.EncodeToString(notebook_file)
	addresses := cacao.Addresses{"url": []string{"http://localhost:7000"}}
	command := cacao.Command{
		Type:       cacao.CommandTypeJupyter,
		CommandB64: notebook_b64,
	}
	jupyter_capability := jupyter.JupyterCapability{}
	context := capability.Context{Command: command, Target: cacao.AgentTarget{Address: addresses}}
	metadata := execution.Metadata{}
	variables, err := jupyter_capability.Execute(metadata, context)
	assert.NotNil(t, err)
	assert.Len(t, variables, 0)
}
