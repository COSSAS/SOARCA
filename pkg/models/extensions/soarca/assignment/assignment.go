package assignment

import "encoding/json"

const AssignmentExtensionType = "soarca-assignment"

type Expression struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

type Assignment struct {
	Type       string     `json:"type"`
	StepResult string     `json:"step-result"`
	Variable   string     `json:"variable"`
	Expression Expression `json:"expression"`
}

func DecodeAssignment(raw interface{}) (Assignment, bool) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return Assignment{}, false
	}
	var model Assignment
	if err := json.Unmarshal(encoded, &model); err != nil {
		return Assignment{}, false
	}
	if model.Type != AssignmentExtensionType {
		return Assignment{}, false
	}
	return model, true
}
