package thehive_test

import (
	"fmt"
	"os"
	"soarca/pkg/integration/thehive/common/connector"
	thehive "soarca/pkg/integration/thehive/reporter"
	"soarca/pkg/models/cacao"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func LoadEnv(key string) (string, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return "", err
	}

	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	} else {
		return "", fmt.Errorf("key: %s not found in .env file", key)
	}
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
	thr := thehive.NewReporter(connector.NewConnector(thehive_api_base_uri, thehive_api_tkn, true))
	str := thr.ConnectorTest()
	fmt.Println(str)
}

func TestTheHiveReporting(t *testing.T) {
	thehive_api_tkn, err := LoadEnv("THEHIVE_TEST_API_TOKEN")
	if err != nil {
		t.Fail()
	}
	thehive_api_base_uri, err := LoadEnv("THEHIVE_TEST_API_BASE_URI")
	if err != nil {
		t.Fail()
	}
	thr := thehive.NewReporter(connector.NewConnector(thehive_api_base_uri, thehive_api_tkn, true))

	expectedCommand := cacao.Command{
		Type:    "ssh",
		Command: "ssh ls -la",
	}

	expectedVariables := cacao.Variable{
		Type:  "string",
		Name:  "var1",
		Value: "testing",
	}
	playbookVariables := cacao.NewVariables(
		cacao.Variable{
			Type:  "string",
			Name:  "__playbook_var__",
			Value: "testing!",
		},
	)

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
		PlaybookVariables:             playbookVariables,
		AuthenticationInfoDefinitions: map[string]cacao.AuthenticationInformation{"id": expectedAuth},
		AgentDefinitions:              map[string]cacao.AgentTarget{"agent1": expectedAgent},
		TargetDefinitions:             map[string]cacao.AgentTarget{"target1": expectedTarget},

		Workflow: map[string]cacao.Step{step1.ID: step1, end.ID: end},
	}
	executionId0 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c0")

	err = thr.ReportWorkflowStart(executionId0, playbook, time.Now())
	if err != nil {
		fmt.Println("failing at report workflow start")
		fmt.Println(err)
		t.Fail()
	}
	err = thr.ReportStepStart(executionId0, step1, cacao.NewVariables(expectedVariables), time.Now())
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	err = thr.ReportStepEnd(executionId0, step1, cacao.NewVariables(expectedVariables), nil, time.Now())
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	err = thr.ReportWorkflowEnd(executionId0, playbook, nil, time.Now())
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}
