package action

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/core/executors"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/reporter"
	timeUtil "soarca/pkg/utils/time"
)

var component = reflect.TypeOf(Executor{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(capabilities map[string]capability.ICapability, reporter reporter.IStepReporter, time timeUtil.ITime) *Executor {
	var instance = Executor{}
	instance.capabilities = capabilities
	instance.reporter = reporter
	instance.time = time
	return &instance
}

type IExecutor interface {
	Execute(metadata execution.Metadata,
		step executors.PlaybookStepMetadata) (cacao.Variables, error)
}

type Executor struct {
	capabilities map[string]capability.ICapability
	reporter     reporter.IStepReporter
	time         timeUtil.ITime
}

type data struct {
	command        cacao.Command
	authentication cacao.AuthenticationInformation
	target         cacao.AgentTarget
	variables      cacao.Variables
	agent          cacao.AgentTarget
}

func (executor *Executor) Execute(meta execution.Metadata,
	metadata executors.PlaybookStepMetadata) (cacao.Variables, error) {

	executor.reporter.ReportStepStart(meta.ExecutionId, metadata.Step, metadata.Variables, executor.time.Now())

	returnVariables := cacao.NewVariables()
	var err error
	defer func() {
		executor.reporter.ReportStepEnd(meta.ExecutionId, metadata.Step, returnVariables, err, executor.time.Now())
	}()

	if metadata.Step.Type != cacao.StepTypeAction {
		err = errors.New("the provided step type is not compatible with this executor")
		log.Error(err)
		return cacao.NewVariables(), err
	}

	returnVariables, err = executor.executeCommandFromArray(meta, metadata)
	return returnVariables, err
}

func (executor *Executor) executeCommandFromArray(meta execution.Metadata,
	metadata executors.PlaybookStepMetadata) (cacao.Variables, error) {
	returnVariables := cacao.NewVariables()
	for _, command := range metadata.Step.Commands {
		// NOTE: This assumes we want to run Command for every Target individually.
		//       Is that something we want to enforce or leave up to the capability?
		for _, element := range metadata.Step.Targets {
			// NOTE: What about Agent authentication?
			target := metadata.Targets[element]
			auth := metadata.Auth[target.AuthInfoIdentifier]

			data := data{
				command:        command,
				authentication: auth,
				target:         target,
				variables:      metadata.Variables,
				agent:          metadata.Agent,
			}

			outputVariables, err := executor.executeCommands(
				meta,
				data)

			if len(metadata.Step.OutArgs) > 0 {
				// If OutArgs is set, only update execution args that are explicitly referenced
				outputVariables = outputVariables.Select(metadata.Step.OutArgs)
			}

			returnVariables.Merge(outputVariables)

			if err != nil {
				log.Error("Error executing Command ", err)
				return cacao.NewVariables(), err
			} else {
				log.Debug("Command executed")
			}
		}
	}
	return returnVariables, nil
}

func interpolateCommand(command cacao.Command, variables cacao.Variables) cacao.Command {
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
	return command
}

func interpolatedTarget(target cacao.AgentTarget, variables cacao.Variables) cacao.AgentTarget {
	for key, addresses := range target.Address {
		var slice []string
		for _, address := range addresses {
			slice = append(slice, variables.Interpolate(address))
		}
		target.Address[key] = slice
	}
	return target
}

func interpolateAuthentication(authentication cacao.AuthenticationInformation, variables cacao.Variables) cacao.AuthenticationInformation {
	authentication.Username = variables.Interpolate(authentication.Username)
	authentication.Password = variables.Interpolate(authentication.Password)
	authentication.UserId = variables.Interpolate(authentication.UserId)
	authentication.Token = variables.Interpolate(authentication.Token)
	authentication.OauthHeader = variables.Interpolate(authentication.OauthHeader)
	authentication.PrivateKey = variables.Interpolate(authentication.PrivateKey)

	return authentication

}

func (executor *Executor) executeCommands(metadata execution.Metadata,
	data data) (cacao.Variables, error) {

	context := capability.Context{}

	if capability, ok := executor.capabilities[data.agent.Name]; ok {
		context.Command = interpolateCommand(data.command, data.variables)
		context.Target = interpolatedTarget(data.target, data.variables)
		context.Authentication = interpolateAuthentication(data.authentication, data.variables)
		context.Variables = data.variables
		returnVariables, err := capability.Execute(metadata, context)
		return returnVariables, err
	} else {
		empty := cacao.NewVariables()
		err := errors.New(fmt.Sprint("capability: ", data.agent.Name, " is not available in soarca"))
		log.Error(err)
		return empty, err
	}

}
