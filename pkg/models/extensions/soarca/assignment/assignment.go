package assignment

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
