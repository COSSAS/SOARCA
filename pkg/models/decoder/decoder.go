package decoder

import (
	"encoding/json"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/validator"
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

func SetPlaybookKeysAsId(playbook *cacao.Playbook) {
	for key, step := range playbook.Workflow {
		step.ID = key
		playbook.Workflow[key] = step
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
}

func decode(data []byte) *cacao.Playbook {
	playbook := cacao.NewPlaybook()

	if err := json.Unmarshal(data, playbook); err != nil {
		log.Error(err)
		return nil
	}

	SetPlaybookKeysAsId(playbook)

	return playbook
}
