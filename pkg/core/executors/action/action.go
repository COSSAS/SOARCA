package action

import (
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/core/executors"
	"soarca/pkg/extensions/soarca/assignment"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	assignmentModel "soarca/pkg/models/extensions/soarca/assignment"
	"soarca/pkg/reporting/reporter"
	timeUtil "soarca/pkg/utils/time"
)

var component = reflect.TypeOf(Executor{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(capabilities map[string]capability.ICapability, reporter reporter.IStepReporter, time timeUtil.ITime, assigner assignment.IAssignmentExtension) *Executor {
	var instance = Executor{}
	instance.capabilities = capabilities
	instance.reporter = reporter
	instance.time = time
	instance.assigner = assigner
	return &instance
}

type IExecuter interface {
	Execute(metadata execution.Metadata,
		step executors.PlaybookStepMetadata) (cacao.Variables, error)
}

type Executor struct {
	capabilities map[string]capability.ICapability
	reporter     reporter.IStepReporter
	time         timeUtil.ITime
	assigner     assignment.IAssignmentExtension
}

type data struct {
	commands  []cacao.Command
	targets   []capability.ResolvedTarget
	variables cacao.Variables
	agent     cacao.AgentTarget
	step      cacao.Step
}

func (executor *Executor) Execute(meta execution.Metadata,
	metadata executors.PlaybookStepMetadata) (cacao.Variables, error) {

	executor.reporter.ReportStepStart(meta, metadata.Step, metadata.Variables, executor.time.Now())

	returnVariables := cacao.NewVariables()
	var err error
	defer func() {
		executor.reporter.ReportStepEnd(meta, metadata.Step, returnVariables, err, executor.time.Now())
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

	// NOTE: interpolation happens once per command/target below, inside
	// executeCommands, so raw (uninterpolated) values are passed through
	// here; only the target->auth resolution needs the raw step data.
	//
	// NOTE: an action step's targets are optional per the CACAO spec — an
	// empty Targets list is a valid shape, not a signal to skip execution.
	// What about Agent authentication? (left as a pre-existing open question)
	targets := make([]capability.ResolvedTarget, 0, len(metadata.Step.Targets))
	for _, element := range metadata.Step.Targets {
		target := metadata.Targets[element]
		auth := metadata.Auth[target.AuthInfoIdentifier]
		targets = append(targets, capability.ResolvedTarget{
			Target:         target,
			Authentication: auth,
		})
	}

	data := data{
		commands:  metadata.Step.Commands,
		targets:   targets,
		variables: metadata.Variables,
		agent:     metadata.Agent,
		step:      metadata.Step,
	}

	outputVariables, err := executor.executeCommands(meta, data)
	if err != nil {
		log.Error("Error executing Command ", err)
		return cacao.NewVariables(), err
	}
	log.Trace("Command executed")

	// Map defined step results into variables as described by any
	// soarca-assignment step extensions.
	assignedVariables := executor.evaluateAssignments(metadata.Step.StepExtensions, outputVariables)
	outputVariables.Merge(assignedVariables)

	if len(metadata.Step.OutArgs) > 0 {
		// If OutArgs is set, only update execution args that are explicitly referenced
		outputVariables = outputVariables.Select(metadata.Step.OutArgs)
	}

	return outputVariables, nil
}

func (executor *Executor) evaluateAssignments(extensions cacao.Extensions, results cacao.Variables) cacao.Variables {
	assigned := cacao.NewVariables()
	for id, raw := range extensions {
		switch raw.(type) {
		case assignmentModel.Assignment:

		}
		model, ok := assignmentModel.DecodeAssignment(raw)
		if !ok {
			continue
		}
		log.Trace("evaluating assignment extension ", id)
		produced := executor.assigner.AssignAndEvaluate(assignment.Context{
			AssignmentModel: model,
			Source:          results,
		})
		assigned.Merge(produced)
	}
	return assigned
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

	if cap, ok := executor.capabilities[data.agent.Type]; ok {
		context := capability.Context{
			Commands:  interpolateCommands(data.commands, data.variables),
			Targets:   interpolateTargets(data.targets, data.variables),
			Variables: data.variables,
			Step:      data.step,
		}
		returnVariables, err := cap.Execute(metadata, context)
		return returnVariables, err
	} else {
		empty := cacao.NewVariables()
		err := errors.New(fmt.Sprint("capability: ", data.agent.Type, " is not available in soarca"))
		log.Error(err)
		return empty, err
	}

}

func interpolateCommands(commands []cacao.Command, variables cacao.Variables) []cacao.Command {
	interpolated := make([]cacao.Command, 0, len(commands))
	for _, command := range commands {
		interpolated = append(interpolated, interpolateCommand(command, variables))
	}
	return interpolated
}

func interpolateTargets(targets []capability.ResolvedTarget, variables cacao.Variables) []capability.ResolvedTarget {
	interpolated := make([]capability.ResolvedTarget, 0, len(targets))
	for _, resolvedTarget := range targets {
		interpolated = append(interpolated, capability.ResolvedTarget{
			Target:         interpolatedTarget(resolvedTarget.Target, variables),
			Authentication: interpolateAuthentication(resolvedTarget.Authentication, variables),
		})
	}
	return interpolated
}
