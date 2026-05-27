package assignment

import (
	"soarca/pkg/models/cacao"
	assignmentModel "soarca/pkg/models/extensions/soarca/assignment"
)

type IAssignmentExtension interface {
	AssignAndEvaluate(Context) cacao.Variables
}

type Context struct {
	AssignmentModel assignmentModel.Assignment
	Source          cacao.Variables
}

type Assignment struct {
}

func (assignment *Assignment) AssignAndEvaluate(context Context) cacao.Variables {

	return nil
}
