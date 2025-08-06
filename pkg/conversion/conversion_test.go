package conversion

import (
	"os"
	"testing"

	"github.com/go-playground/assert/v2"
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
	ssh_simple_file, err := os.ReadFile("../../test/conversion/simple_ssh.bpmn")
	assert.Equal(t, err, nil)
	converted, err := PerformConversion("../../test/conversion/simple_ssh.bpmn", ssh_simple_file, "bpmn")
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
	ssh_simple_file, err := os.ReadFile("../../test/conversion/control_gates.bpmn")
	assert.Equal(t, err, nil)
	converted, err := PerformConversion("../../test/conversion/control_gates.bpmn", ssh_simple_file, "bpmn")
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
