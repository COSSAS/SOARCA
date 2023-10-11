package cacao_test

import (
	"fmt"
	"io"
	"os"
	"soarca/models/cacao"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func getTime(data string) time.Time {
	res, _ := time.Parse(time.RFC3339, data)
	return res
}

func TestCacaoDecode(t *testing.T) {
	// p := cacao.Playbook{}
	// fmt.Println(p)

	jsonFile, err := os.Open("playbook.json")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Not valid JSON")
		t.Fail()
		return
	}

	var workflow = cacao.Decode(byteValue)

	// fmt.Println(workflow)

	// for _, w := range workflow.Workflow {
	// 	fmt.Println(w.ID)
	// 	for _, c := range w.Commands {
	// 		fmt.Println(c.Type)
	// 	}
	// }
	assert.Equal(t, workflow.ID, "playbook--61a6c41e-6efc-4516-a242-dfbc5c89d562")
	assert.Equal(t, workflow.Type, "playbook")
	assert.Equal(t, workflow.SpecVersion, "cacao-2.0")
	assert.Equal(t, workflow.Name, "SOARCA Main Flow")
	assert.Equal(t, workflow.Description, "This playbook will run for each trigger event in SOARCA")
	assert.Equal(t, workflow.PlaybookTypes, []string{"notification"})
	assert.Equal(t, workflow.CreatedBy, "identity--5abe695c-7bd5-4c31-8824-2528696cdbf1")

	assert.Equal(t, workflow.Created, getTime("2023-05-26T15:56:00.123456Z"))
	assert.Equal(t, workflow.Modified, getTime("2023-05-26T15:56:00.123456Z"))
	assert.Equal(t, workflow.ValidFrom, getTime("2023-05-26T15:56:00.123456Z"))
	assert.Equal(t, workflow.ValidUntil, getTime("2337-05-26T15:56:00.123456Z"))
	assert.Equal(t, workflow.Priority, 1)
	assert.Equal(t, workflow.Severity, 1)
	assert.Equal(t, workflow.Impact, 1)
	assert.Equal(t, workflow.Labels, []string{"soarca"})
	assert.Equal(t, len(workflow.ExternalReferences), 1)
	assert.Equal(t, workflow.ExternalReferences[0].Name, "TNO SOARCA")
	assert.Equal(t, workflow.ExternalReferences[0].Description, "SOARCA Homepage")
	assert.Equal(t, workflow.ExternalReferences[0].Source, "TNO CST")
	assert.Equal(t, workflow.ExternalReferences[0].URL, "http://tno.nl/cst")
	assert.Equal(t, workflow.WorkflowStart, "step--a76dbc32-b739-427b-ae13-4ec703d5797e")
	assert.Equal(t, workflow.WorkflowException, "step--40131926-89e9-44df-a018-5f92f2df7914")

	assert.Equal(t, len(workflow.AuthenticationInfoDefinitions), 1)
	assert.Equal(t, workflow.AuthenticationInfoDefinitions["http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae"].UserId, "admin")
	assert.Equal(t, workflow.AuthenticationInfoDefinitions["http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae"].Password, "super-secure-password")

	assert.Equal(t, len(workflow.Workflow), 5)
	step1 := workflow.Workflow["step--a76dbc32-b739-427b-ae13-4ec703d5797e"]
	assert.Equal(t, step1.ID, "step--a76dbc32-b739-427b-ae13-4ec703d5797e")
	assert.Equal(t, step1.ObjectType, "action")
	assert.Equal(t, step1.Name, "IMC assets by CVE")
	assert.Equal(t, step1.Description, "Check the IMC for affected assets by CVE")
	assert.Equal(t, step1.OnCompletion, "step--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4")
	assert.Equal(t, len(step1.Commands), 1)
	assert.Equal(t, step1.Commands[0].Command, "GET http://__imc_address__/by/__cve__")
	assert.Equal(t, step1.Commands[0].Type, "http-api")

	step2 := workflow.Workflow["step--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4"]
	assert.Equal(t, step2.ID, "step--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4")
	assert.Equal(t, step2.ObjectType, "action")
	assert.Equal(t, step2.Name, "BIA for CVE")
	assert.Equal(t, step2.Description, "Perform Business Impact Analysis for CVE")
	assert.Equal(t, step2.OnCompletion, "step--09b97fab-56a1-45dc-a88f-be3cde3eac33")
	assert.Equal(t, len(step2.Commands), 1)
	assert.Equal(t, step2.Commands[0].Command, "GET http://__bia_address__/analysisreport/__cve__")
	assert.Equal(t, step2.Commands[0].Type, "http-api")

	step3 := workflow.Workflow["step--09b97fab-56a1-45dc-a88f-be3cde3eac33"]
	assert.Equal(t, step3.ID, "step--09b97fab-56a1-45dc-a88f-be3cde3eac33")
	assert.Equal(t, step3.ObjectType, "action")
	assert.Equal(t, step3.Name, "Generate CoAs")
	assert.Equal(t, step3.Description, "Generate Courses of Action")
	assert.Equal(t, step3.OnCompletion, "step--2190f685-1857-44ac-ad0e-0ded6c6ef3ce")
	assert.Equal(t, len(step3.Commands), 1)
	assert.Equal(t, step3.Commands[0].Command, "GET http://__coagenerator_address__/coa/__assetuuid__")
	assert.Equal(t, step3.Commands[0].Type, "http-api")

	step4 := workflow.Workflow["step--2190f685-1857-44ac-ad0e-0ded6c6ef3ce"]
	assert.Equal(t, step4.ID, "step--2190f685-1857-44ac-ad0e-0ded6c6ef3ce")
	assert.Equal(t, step4.ObjectType, "action")
	assert.Equal(t, step4.Name, "BIA for CoAs")
	assert.Equal(t, step4.Description, "Perform Business Impact Analysis for Courses of Action")
	assert.Equal(t, step4.OnCompletion, "end--6b23c237-ade8-4d00-9aa1-75999738d557")
	assert.Equal(t, len(step4.Commands), 1)
	assert.Equal(t, step4.Commands[0].Command, "GET http://__bia_address__/analysisreport/__coa_list__")
	assert.Equal(t, step4.Commands[0].Type, "http-api")

	step5 := workflow.Workflow["end--6b23c237-ade8-4d00-9aa1-75999738d557"]
	assert.Equal(t, step5.ID, "end--6b23c237-ade8-4d00-9aa1-75999738d557")
	assert.Equal(t, step5.ObjectType, "end")
	assert.Equal(t, step5.Name, "End SOARCA Main Flow")

}
