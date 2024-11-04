package thehive_test

import (
	"bufio"
	"fmt"
	"os"
	"soarca/internal/reporter/downstream_reporter/thehive"
	"soarca/internal/reporter/downstream_reporter/thehive/connector"
	"soarca/models/cacao"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Microsoft Copilot provided code to get .env local file and extract variables values
func LoadEnv(envVar string) (string, error) {
	file, err := os.Open(".env")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, envVar+"=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(parts[1], `"`), nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("variable %s not found", envVar)
}

func TestTheHiveConnection(t *testing.T) {
	thehive_api_tkn, err := LoadEnv("THEHIVE_TEST_API_TOKEN")
	if err != nil {
		t.Fail()
	}
	thehive_api_base_uri, err := LoadEnv("THEHIVE_TEST_API_BASE_URI")
	if err != nil {
		t.Fail()
	}
	thr := thehive.New(connector.New(thehive_api_base_uri, thehive_api_tkn))
	str := thr.ConnectorTest()
	fmt.Println(str)
}

func TestTheHiveOpenCase(t *testing.T) {
	thehive_api_tkn, err := LoadEnv("THEHIVE_TEST_API_TOKEN")
	if err != nil {
		t.Fail()
	}
	thehive_api_base_uri, err := LoadEnv("THEHIVE_TEST_API_BASE_URI")
	if err != nil {
		t.Fail()
	}
	thr := thehive.New(connector.New(thehive_api_base_uri, thehive_api_tkn))

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}

	step1 := cacao.Step{
		Type:          "action",
		ID:            "action--test",
		Name:          "ssh-tests",
		Description:   "test step",
		StepVariables: cacao.NewVariables(expectedVariables),
		Commands:      []cacao.Command{expectedCommand},
		Cases:         map[string]string{},
		OnCompletion:  "end--test",
		Agent:         "agent1",
		Targets:       []string{"target1"},
	}

	end := cacao.Step{
		Type: "end",
		ID:   "end--test",
		Name: "end step",
	}

	expectedAuth := cacao.AuthenticationInformation{
		Name: "user",
		ID:   "auth1",
	}

	expectedTarget := cacao.AgentTarget{
		Name:               "sometarget",
		AuthInfoIdentifier: "auth1",
		ID:                 "target1",
	}

	expectedAgent := cacao.AgentTarget{
		Type: "soarca",
		Name: "soarca-ssh",
	}

	playbook := cacao.Playbook{
		ID:                            "test",
		Type:                          "test",
		Name:                          "ssh-test-playbook",
		Description:                   "Playbook description",
		WorkflowStart:                 step1.ID,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")

	err = thr.ReportWorkflowStart(executionId0, playbook)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	// err = thr.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables))
	// if err != nil {
	// 	fmt.Println(err)
	// 	t.Fail()
	// }
}
