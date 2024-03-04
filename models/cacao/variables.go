package cacao

import (
	"fmt"
	"strings"
)

type Variable struct {
	Type        string `bson:"type" json:"type" validate:"required"`
	Name        string `bson:"name,omitempty" json:"name,omitempty"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Value       string `bson:"value,omitempty" json:"value,omitempty"`
	Constant    bool   `bson:"constant,omitempty" json:"constant,omitempty"`
	External    bool   `bson:"external,omitempty" json:"external,omitempty"`
}

type VariableMap map[string]Variable

func (variableMap VariableMap) Merge(otherMap VariableMap) VariableMap {
	newMap := make(VariableMap)
	for k, v := range variableMap {
		newMap[k] = v
	}

	for k, v := range otherMap {
		if _, ok := variableMap[k]; !ok || v.Constant {
			newMap[k] = v
		}
	}

	return newMap
}

// TODO: Find a way to store the Replacer 'inside' the VariableMap
func (variableMap VariableMap) Replace(s string) string {
	replacements := make([]string, 0)
	for k, v := range variableMap {
		replacementKey := fmt.Sprintf("%s:value", k)
		replacements = append(replacements, replacementKey, v.Value)
	}
	r := strings.NewReplacer(replacements...)
	return r.Replace(s)
}
