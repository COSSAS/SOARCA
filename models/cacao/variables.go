package cacao

import (
	"fmt"
	"strings"
)

// Merge two maps of Cacao variables
//
// Existing values are not overwritten unless the replacement has been
// marked as constant.
func (variableMap VariableMap) Merge(otherMap VariableMap) VariableMap {
	newMap := make(VariableMap)
	for key, value := range variableMap {
		newMap[key] = value
	}

	for key, value := range otherMap {
		if _, ok := variableMap[key]; !ok || value.Constant {
			newMap[key] = value
		}
	}

	return newMap
}

// Replace variable references in a string
//
// Variable substitution is performed according to the CACAO spec,
// which states `__var__:value` is replaced with the value of `__var__`.
func (variableMap VariableMap) Replace(s string) string {
	replacements := make([]string, 0)
	for key, value := range variableMap {
		replacementKey := fmt.Sprintf("%s:value", key)
		replacements = append(replacements, replacementKey, value.Value)
	}
	replacer := strings.NewReplacer(replacements...)
	return replacer.Replace(s)
}

