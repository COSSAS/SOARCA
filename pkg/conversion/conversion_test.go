package conversion

import (
	"encoding/json"
	"os"
	"soarca/pkg/models/validator"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_read_format(t *testing.T) {
	assert.Equal(t, read_format("bpmn"), FormatBpmn)
	assert.Equal(t, read_format("splunk"), FormatSplunk)
	assert.Equal(t, read_format("misp"), FormatMisp)
	assert.Equal(t, read_format(""), FormatUnknown)
	assert.Equal(t, read_format("cacao"), FormatUnknown)
	assert.Equal(t, read_format("bpnm"), FormatUnknown)
	assert.Equal(t, read_format("?"), FormatUnknown)
}
func Test_guess_format(t *testing.T) {
	assert.Equal(t, guess_format("x.bpmn"), FormatBpmn)
}

func Test_bpmn_format(t *testing.T) {
	content, err := os.ReadFile("../../test/conversion/simple_ssh.bpmn")
	assert.Equal(t, err, nil)
	converted, err := PerformConversion("../../test/conversion/simple_ssh.bpmn", content, "bpmn")
	assert.Equal(t, err, nil)
	converted_json, err := json.Marshal(converted)
	assert.Nil(t, err)
	err = validator.IsValidCacaoJson(converted_json)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, converted, nil)
	assert.MatchRegex(t, converted.WorkflowStart, "start--.*")
	assert.MatchRegex(t, converted.WorkflowException, "end--.*")
	assert.NotEqual(t, converted.Workflow, nil)
	for _, entry := range converted.Workflow {
		assert.NotEqual(t, entry.Name, nil)
		assert.NotEqual(t, entry.Type, nil)
	}
	assert.Equal(t, len(converted.Workflow), 4)
}
func Test_bpmn_format_control(t *testing.T) {
	content, err := os.ReadFile("../../test/conversion/control_gates.bpmn")
	assert.Equal(t, err, nil)
	converted, err := PerformConversion("../../test/conversion/control_gates.bpmn", content, "bpmn")
	assert.Equal(t, err, nil)
	converted_json, err := json.Marshal(converted)
	err = validator.IsValidCacaoJson(converted_json)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, converted, nil)
	assert.MatchRegex(t, converted.WorkflowStart, "start--.*")
	assert.MatchRegex(t, converted.WorkflowException, "end--.*")
	assert.NotEqual(t, converted.Workflow, nil)
	for _, entry := range converted.Workflow {
		assert.NotEqual(t, entry.Name, nil)
		assert.NotEqual(t, entry.Type, nil)
	}
	assert.Equal(t, len(converted.Workflow), 11)
}
