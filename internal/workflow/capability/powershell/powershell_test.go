package powershell

import (
	"soarca/internal/workflow/capability"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestPowershellExecuteNoTargetsSkipsWithoutError(t *testing.T) {
	powershellCapability := New()

	runId, _ := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	playbookId, _ := uuid.Parse("d09351a2-a075-40c8-8054-0b7c423db83f")
	stepId, _ := uuid.Parse("81eff59f-d084-4324-9e0a-59e353dbd28f")

	command := cacao.Command{Type: "powershell", Command: "Get-Process"}

	metadata := run.Metadata{
		RunId:      runId,
		PlaybookId: playbookId.String(),
		StepId:     stepId.String(),
	}

	data := capability.Context{
		Commands: []cacao.Command{command},
		Targets:  []capability.ResolvedTarget{},
	}

	results, err := powershellCapability.Execute(metadata, data)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, len(results), 0)
}
