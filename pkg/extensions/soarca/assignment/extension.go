package assignment

import (
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/cacao"
	"soarca/pkg/extensions/soarca/assignment/expression"
	"soarca/pkg/extensions/soarca/assignment/expression/jq"
	"soarca/pkg/extensions/soarca/assignment/expression/regex"
)

type Empty struct{}

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type IAssignmentExtension interface {
	AssignAndEvaluate(Context) cacao.Variables
}

type Context struct {
	AssignmentModel Assignment
	Source          cacao.Variables
}

// Evaluator applies assignment extension definitions to step results.
type Evaluator struct {
	engines map[string]expression.IExpression
}

// New returns an Evaluator with the default set of expression engines
// registered, keyed on the engine name as it appears in an expression's
// "type" field.
func New() *Evaluator {
	jqEngine := jq.New()
	regexEngine := regex.New()
	return &Evaluator{
		engines: map[string]expression.IExpression{
			jqEngine.GetEngineName():    jqEngine,
			regexEngine.GetEngineName(): regexEngine,
		},
	}
}

// AssignAndEvaluate maps a single defined step result into a target variable.
// When the assignment carries an expression, the step result is first run
// through the matching engine. The returned Variables contain the target
// variable on success, or are empty when the assignment cannot be applied
// (the reason is logged).
func (assignment *Evaluator) AssignAndEvaluate(context Context) cacao.Variables {
	variables := cacao.NewVariables()
	model := context.AssignmentModel

	if model.StepResult == "" || model.Variable == "" {
		log.Warning("assignment is missing a step-result or variable, skipping assignment")
		return variables
	}

	source, found := context.Source.Find(model.StepResult)
	if !found {
		log.Warningf("step result %q is not present in the step output, skipping assignment", model.StepResult)
		return variables
	}

	value := source.Value

	// An empty expression type means the raw step result is mapped as-is.
	if model.Expression.Type != "" {
		engine, ok := assignment.engines[model.Expression.Type]
		if !ok {
			log.Warningf("expression engine %q is not supported, skipping assignment", model.Expression.Type)
			return variables
		}
		evaluated, err := engine.Execute(value, expression.Query(model.Expression.Expression))
		if err != nil {
			log.Warningf("%s expression failed, skipping assignment: %s", model.Expression.Type, err.Error())
			return variables
		}
		value = evaluated
	}

	variables.InsertOrReplace(cacao.Variable{
		Type:  cacao.VariableTypeString,
		Name:  model.Variable,
		Value: value,
	})
	return variables
}
