package assignment

import (
	"testing"

	"soarca/pkg/models/cacao"
	assignmentModel "soarca/pkg/models/extensions/soarca/assignment"

	"github.com/go-playground/assert/v2"
)

const httpResultName = "__soarca_http_api_result__"

func source(body string) cacao.Variables {
	return cacao.NewVariables(cacao.Variable{
		Type:  cacao.VariableTypeString,
		Name:  httpResultName,
		Value: body,
	})
}

func TestAssignPassthrough(t *testing.T) {
	body := `{"status": 200, "body": "ok"}`
	out := New().AssignAndEvaluate(Context{
		AssignmentModel: assignmentModel.Assignment{
			Type:       "soarca-assignment",
			StepResult: httpResultName,
			Variable:   "__raw_result__",
		},
		Source: source(body),
	})

	variable, found := out.Find("__raw_result__")
	assert.Equal(t, found, true)
	assert.Equal(t, variable.Value, body)
	assert.Equal(t, variable.Type, cacao.VariableTypeString)
	assert.Equal(t, len(out), 1)
}

func TestAssignWithJqExpression(t *testing.T) {
	body := `{"status": 200, "headers": {"location": "https://example.test/here"}}`
	out := New().AssignAndEvaluate(Context{
		AssignmentModel: assignmentModel.Assignment{
			Type:       "soarca-assignment",
			StepResult: httpResultName,
			Variable:   "__location__",
			Expression: assignmentModel.Expression{Type: "jq", Expression: ".headers.location"},
		},
		Source: source(body),
	})

	variable, found := out.Find("__location__")
	assert.Equal(t, found, true)
	assert.Equal(t, variable.Value, "https://example.test/here")
}

func TestAssignMissingStepResult(t *testing.T) {
	out := New().AssignAndEvaluate(Context{
		AssignmentModel: assignmentModel.Assignment{StepResult: "__does_not_exist__", Variable: "__out__"},
		Source:          source(`{}`),
	})
	assert.Equal(t, len(out), 0)
}

func TestAssignUnknownEngine(t *testing.T) {
	out := New().AssignAndEvaluate(Context{
		AssignmentModel: assignmentModel.Assignment{
			StepResult: httpResultName,
			Variable:   "__out__",
			Expression: assignmentModel.Expression{Type: "sed", Expression: "s/a/b/"},
		},
		Source: source(`{}`),
	})
	assert.Equal(t, len(out), 0)
}

func TestAssignJqError(t *testing.T) {
	out := New().AssignAndEvaluate(Context{
		AssignmentModel: assignmentModel.Assignment{
			StepResult: httpResultName,
			Variable:   "__out__",
			Expression: assignmentModel.Expression{Type: "jq", Expression: ".headers["},
		},
		Source: source(`{"headers": {}}`),
	})
	assert.Equal(t, len(out), 0)
}

func TestAssignMissingVariableName(t *testing.T) {
	out := New().AssignAndEvaluate(Context{
		AssignmentModel: assignmentModel.Assignment{StepResult: httpResultName, Variable: ""},
		Source:          source(`{}`),
	})
	assert.Equal(t, len(out), 0)
}
