package caldera

import (
	"fmt"
	"github.com/go-playground/assert/v2"
	"soarca/pkg/core/capability/caldera"
	"soarca/models/cacao"
	"soarca/models/execution"
	"testing"
)

func TestCapabilityName(t *testing.T) {
	calderaCapability := caldera.New()
	assert.Equal(t, calderaCapability.GetType(), "soarca-caldera-cmd")
}

func TestExecute(t *testing.T) {
	calderaCapability := caldera.New()

	results, err := calderaCapability.Execute(
		execution.Metadata{},
		cacao.Command{},
		cacao.AuthenticationInformation{},
		cacao.AgentTarget{},
		cacao.NewVariables())

	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	assert.Equal(t, cacao.NewVariables(), results)
}
