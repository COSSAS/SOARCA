package action

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/capability"
	"soarca/internal/reporter"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
)

var component = reflect.TypeOf(Executor{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(capabilities map[string]capability.ICapability, reporter reporter.IStepReporter) *Executor {
	var instance = Executor{}
	instance.capabilities = capabilities
	instance.reporter = reporter
	return &instance
}

type PlaybookStepMetadata struct {
	Step      cacao.Step
	Targets   map[string]cacao.AgentTarget
	Auth      map[string]cacao.AuthenticationInformation
	Agent     cacao.AgentTarget
	Variables cacao.Variables
}

type IExecuter interface {
	Execute(metadata execution.Metadata,
		step PlaybookStepMetadata) (cacao.Variables, error)
}

type Executor struct {
	capabilities map[string]capability.ICapability
	reporter     reporter.IStepReporter
}

func (executor *Executor) Execute(meta execution.Metadata, metadata PlaybookStepMetadata) (cacao.Variables, error) {

	executor.reporter.ReportStepStart(meta.ExecutionId, metadata.Step, metadata.Variables)

	if metadata.Step.Type != cacao.StepTypeAction {
		err := errors.New("the provided step type is not compatible with this executor")
		log.Error(err)
		executor.reporter.ReportStepEnd(meta.ExecutionId, metadata.Step, cacao.NewVariables(), err)
		return cacao.NewVariables(), err
	}
	returnVariables := cacao.NewVariables()
	for _, command := range metadata.Step.Commands {
		// NOTE: This assumes we want to run Command for every Target individually.
		//       Is that something we want to enforce or leave up to the capability?
		for _, element := range metadata.Step.Targets {
			// NOTE: What about Agent authentication?
			target := metadata.Targets[element]
			auth := metadata.Auth[target.AuthInfoIdentifier]

			outputVariables, err := executor.ExecuteActionStep(
				meta,
				command,
				auth,
				target,
				metadata.Variables,
				metadata.Step,
				metadata.Agent)

			returnVariables.Merge(outputVariables)

			if err != nil {
				log.Error("Error executing Command ", err)
				executor.reporter.ReportStepEnd(meta.ExecutionId, metadata.Step, returnVariables, err)
				return cacao.NewVariables(), err
			} else {
				log.Debug("Command executed")
			}
		}
	}
	executor.reporter.ReportStepEnd(meta.ExecutionId, metadata.Step, returnVariables, nil)
	if len(metadata.Step.OutArgs) > 0 {
		// If OutArgs is set, only update execution args that are explicitly referenced
		returnVariables = returnVariables.Select(metadata.Step.OutArgs)
	}
	return returnVariables, nil
}

func (executor *Executor) ExecuteActionStep(metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.Variables,
	step cacao.Step,
	agent cacao.AgentTarget) (cacao.Variables, error) {

	if capability, ok := executor.capabilities[agent.Name]; ok {
		command.Command = variables.Interpolate(command.Command)
		command.Content = variables.Interpolate(command.Content)
		command.ContentB64 = variables.Interpolate(command.ContentB64)

		for key, headers := range command.Headers {
			var slice []string
			for _, header := range headers {
				slice = append(slice, variables.Interpolate(header))
			}
			command.Headers[key] = slice
		}

		for key, addresses := range target.Address {
			var slice []string
			for _, address := range addresses {
				slice = append(slice, variables.Interpolate(address))
			}
			target.Address[key] = slice
		}

		authentication.Password = variables.Interpolate(authentication.Password)
		authentication.Username = variables.Interpolate(authentication.Username)
		authentication.UserId = variables.Interpolate(authentication.UserId)
		authentication.Token = variables.Interpolate(authentication.Token)
		authentication.OauthHeader = variables.Interpolate(authentication.OauthHeader)
		authentication.PrivateKey = variables.Interpolate(authentication.PrivateKey)

		returnVariables, err := capability.Execute(metadata,
			command,
			authentication,
			target,
			variables,
			step.InArgs,
			step.OutArgs)

		return returnVariables, err
	} else {
		empty := cacao.NewVariables()
		err := errors.New(fmt.Sprint("capability: ", agent.Name, " is not available in soarca"))
		log.Error(err)
		return empty, err
	}

}
