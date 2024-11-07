package cacao_test

import (
	"fmt"
	"io"
	"os"
	"soarca/models/decoder"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

// NOTE: using CACAO V2 CDS01 SCHEMA because schema for CDS03 has bug:
// https://raw.githubusercontent.com/cyentific-rni/cacao-json-schemas/cacao-v2.0-csd03/schemas/data-markings/string returned status code 404
// The schemas are CDS01 compatible as they have the following properties renamed:
//   - "agents" from CDS01 instead of "agent_definitions" from CDS03+
//   - "targets" from CDS01 instead of "target_definitions" from CDS03+
var PB_PATH string = "playbooks/"

func getTime(data string) time.Time {
	res, _ := time.Parse(time.RFC3339, data)
	return res
}

func TestCacaoDecode(t *testing.T) {
	// p := cacao.Playbook{}
	// fmt.Println(p)

	jsonFile, err := os.Open(PB_PATH + "playbook.json")
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

	var workflow = decoder.DecodeValidate(byteValue)

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
	assert.Equal(t, workflow.WorkflowStart, "action--a76dbc32-b739-427b-ae13-4ec703d5797e")
	assert.Equal(t, workflow.WorkflowException, "end--40131926-89e9-44df-a018-5f92f2df7914")

	assert.Equal(t, len(workflow.AuthenticationInfoDefinitions), 1)
	assert.Equal(t, workflow.AuthenticationInfoDefinitions["http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae"].UserId, "admin")
	assert.Equal(t, workflow.AuthenticationInfoDefinitions["http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae"].Password, "super-secure-password")

	assert.Equal(t, len(workflow.Workflow), 5)
	step1 := workflow.Workflow["action--a76dbc32-b739-427b-ae13-4ec703d5797e"]
	assert.Equal(t, step1.ID, "action--a76dbc32-b739-427b-ae13-4ec703d5797e")
	assert.Equal(t, step1.Type, "action")
	assert.Equal(t, step1.Name, "IMC assets by CVE")
	assert.Equal(t, step1.Description, "Check the IMC for affected assets by CVE")
	assert.Equal(t, step1.OnCompletion, "action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4")
	assert.Equal(t, len(step1.Commands), 1)
	assert.Equal(t, step1.Commands[0].Command, "GET http://__imc_address__/by/__cve__")
	assert.Equal(t, step1.Commands[0].Type, "http-api")

	step2 := workflow.Workflow["action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4"]
	assert.Equal(t, step2.ID, "action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4")
	assert.Equal(t, step2.Type, "action")
	assert.Equal(t, step2.Name, "BIA for CVE")
	assert.Equal(t, step2.Description, "Perform Business Impact Analysis for CVE")
	assert.Equal(t, step2.OnCompletion, "action--09b97fab-56a1-45dc-a88f-be3cde3eac33")
	assert.Equal(t, len(step2.Commands), 1)
	assert.Equal(t, step2.Commands[0].Command, "GET http://__bia_address__/analysisreport/__cve__")
	assert.Equal(t, step2.Commands[0].Type, "http-api")

	step3 := workflow.Workflow["action--09b97fab-56a1-45dc-a88f-be3cde3eac33"]
	assert.Equal(t, step3.ID, "action--09b97fab-56a1-45dc-a88f-be3cde3eac33")
	assert.Equal(t, step3.Type, "action")
	assert.Equal(t, step3.Name, "Generate CoAs")
	assert.Equal(t, step3.Description, "Generate Courses of Action")
	assert.Equal(t, step3.OnCompletion, "action--2190f685-1857-44ac-ad0e-0ded6c6ef3ce")
	assert.Equal(t, len(step3.Commands), 1)
	assert.Equal(t, step3.Commands[0].Command, "GET http://__coagenerator_address__/coa/__assetuuid__")
	assert.Equal(t, step3.Commands[0].Type, "http-api")

	step4 := workflow.Workflow["action--2190f685-1857-44ac-ad0e-0ded6c6ef3ce"]
	assert.Equal(t, step4.ID, "action--2190f685-1857-44ac-ad0e-0ded6c6ef3ce")
	assert.Equal(t, step4.Type, "action")
	assert.Equal(t, step4.Name, "BIA for CoAs")
	assert.Equal(t, step4.Description, "Perform Business Impact Analysis for Courses of Action")
	assert.Equal(t, step4.OnCompletion, "end--6b23c237-ade8-4d00-9aa1-75999738d557")
	assert.Equal(t, len(step4.Commands), 1)
	assert.Equal(t, step4.Commands[0].Command, "GET http://__bia_address__/analysisreport/__coa_list__")
	assert.Equal(t, step4.Commands[0].Type, "http-api")

	step5 := workflow.Workflow["end--6b23c237-ade8-4d00-9aa1-75999738d557"]
	assert.Equal(t, step5.ID, "end--6b23c237-ade8-4d00-9aa1-75999738d557")
	assert.Equal(t, step5.Type, "end")
	assert.Equal(t, step5.Name, "End SOARCA Main Flow")

	assert.Equal(t, workflow.AgentDefinitions["http-api--7e9174ec-a293-43df-a72d-471c79e276bf"].Name, "Firewall 1")
	assert.Equal(t, workflow.AgentDefinitions["http-api--7e9174ec-a293-43df-a72d-471c79e276bf"].ID, "http-api--7e9174ec-a293-43df-a72d-471c79e276bf")
	assert.Equal(t, workflow.AgentDefinitions["http-api--7e9174ec-a293-43df-a72d-471c79e276bf"].Type, "http-api")
	assert.Equal(t, workflow.AgentDefinitions["http-api--7e9174ec-a293-43df-a72d-471c79e276bf"].Address["dname"][0], "hxxp://example.com/v1/")
	assert.Equal(t, workflow.AgentDefinitions["http-api--7e9174ec-a293-43df-a72d-471c79e276bf"].Location.Name, "Eindhoven")

	assert.Equal(t, workflow.AuthenticationInfoDefinitions["http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae"].ID, "http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae")
	assert.Equal(t, workflow.AuthenticationInfoDefinitions["http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae"].Type, "http-basic")
	assert.Equal(t, workflow.AuthenticationInfoDefinitions["http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae"].UserId, "admin")
	assert.Equal(t, workflow.AuthenticationInfoDefinitions["http-basic--76c26f7f-9a15-40ff-a90a-7b19e23372ae"].Password, "super-secure-password")

	// Assert the variables are mapped correctly on the playbook level
	assert.Equal(t, workflow.PlaybookVariables["__var1__"].Name, "__var1__")
	assert.Equal(t, workflow.PlaybookVariables["__var1__"].Type, "string")
	assert.Equal(t, workflow.PlaybookVariables["__var1__"].Description, "Some nice description")
	assert.Equal(t, workflow.PlaybookVariables["__var1__"].Value, "A random string")
	assert.Equal(t, workflow.PlaybookVariables["__var1__"].Constant, false)
	assert.Equal(t, workflow.PlaybookVariables["__var1__"].External, false)

	// Assert the variables are mapped correctly on the step level
	assert.Equal(t, workflow.Workflow["action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4"].StepVariables["__bia_address__"].Name, "__bia_address__")
	assert.Equal(t, workflow.Workflow["action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4"].StepVariables["__bia_address__"].Type, "ipv4-addr")
	assert.Equal(t, workflow.Workflow["action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4"].StepVariables["__bia_address__"].Description, "Bia address")
	assert.Equal(t, workflow.Workflow["action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4"].StepVariables["__bia_address__"].Value, "192.168.0.1")
	assert.Equal(t, workflow.Workflow["action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4"].StepVariables["__bia_address__"].Constant, true)
	assert.Equal(t, workflow.Workflow["action--9fcc5c3b-0b70-4d73-b922-cf5491dcd1a4"].StepVariables["__bia_address__"].External, false)

}
