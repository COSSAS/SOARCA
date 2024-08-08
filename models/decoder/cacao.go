package decoder

import (
	"encoding/json"
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/validator"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func DecodeValidate(data []byte) *cacao.Playbook {

	if err := validator.IsValidCacaoJson(data); err != nil {
		log.Error(err)
		log.Trace("json validation failed")
		return nil
	}

	playbook := decode(data)

	if playbook == nil {
		log.Error("playbook decoding failed")
		return nil
	}

	if err := validator.IsSafeCacaoWorkflow(playbook); err != nil {
		log.Error(err)
		return nil
	}

	return playbook
}

func decode(data []byte) *cacao.Playbook {
	playbook := cacao.NewPlaybook()

	if err := json.Unmarshal(data, playbook); err != nil {
		log.Error(err)
		return nil
	}

	for key, workflow := range playbook.Workflow {
		workflow.ID = key
		playbook.Workflow[key] = workflow
	}

	for key, target := range playbook.TargetDefinitions {
		target.ID = key
		playbook.TargetDefinitions[key] = target
	}

	for key, agent := range playbook.AgentDefinitions {
		agent.ID = key
		playbook.AgentDefinitions[key] = agent
	}

	for key, auth := range playbook.AuthenticationInfoDefinitions {
		auth.ID = key
		playbook.AuthenticationInfoDefinitions[key] = auth
	}

	for key, variable := range playbook.PlaybookVariables {
		variable.Name = key
		playbook.PlaybookVariables.InsertOrReplace(variable)
	}

	for _, step := range playbook.Workflow {
		for key, variable := range step.StepVariables {
			variable.Name = key
			step.StepVariables.InsertOrReplace(variable)
		}
	}

	return playbook
}
