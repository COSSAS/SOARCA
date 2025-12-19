package conversion

import (
	"encoding/json"
	"os"
	model "soarca/pkg/models/conversion"
	"soarca/pkg/models/validator"
	util "soarca/pkg/utils/conversion"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_read_format(t *testing.T) {
	assert.Equal(t, util.ReadFormat("bpmn"), model.FormatBpmn)
	assert.Equal(t, util.ReadFormat(""), model.FormatUnknown)
	assert.Equal(t, util.ReadFormat("cacao"), model.FormatUnknown)
	assert.Equal(t, util.ReadFormat("bpnm"), model.FormatUnknown)
	assert.Equal(t, util.ReadFormat("?"), model.FormatUnknown)
}
func Test_guess_format(t *testing.T) {
	assert.Equal(t, util.GuessFormat("x.bpmn"), model.FormatBpmn)
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
	assert.True(t, strings.HasPrefix(converted.WorkflowStart, "start--"))
	assert.True(t, strings.HasPrefix(converted.WorkflowException, "end--"))
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
	assert.Equal(t, err, nil)
	err = validator.IsValidCacaoJson(converted_json)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, converted, nil)
	assert.True(t, strings.HasPrefix(converted.WorkflowStart, "start--"))
	assert.True(t, strings.HasPrefix(converted.WorkflowException, "end--"))
	assert.NotEqual(t, converted.Workflow, nil)
	for _, entry := range converted.Workflow {
		assert.NotEqual(t, entry.Name, nil)
		assert.NotEqual(t, entry.Type, nil)
	}
	assert.Equal(t, len(converted.Workflow), 11)
}
