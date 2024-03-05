package cacao

import (
	"fmt"
	"strings"
)

// Merge two maps of Cacao variables
//
// Existing values are not overwritten unless the replacement has been
// marked as constant.
func (variables Variables) Merge(otherVariables Variables) Variables {
	newVariables := make(Variables)
	for key, value := range variables {
		newVariables[key] = value
	}

	for key, newValue := range otherVariables {
		if _, ok := variables[key]; !ok || newValue.Constant {
			newVariables[key] = newValue
		}
	}

	return newVariables
}

// Construct a Replacer for string interpolation
//
// Variable substitution is performed according to the CACAO spec,
// which states `__var__:value` is replaced with the value of `__var__`.
func (variables Variables) Replacer() *strings.Replacer {
	replacements := make([]string, 0)
	for key, value := range variables {
		replacementKey := fmt.Sprintf("%s:value", key)
		replacements = append(replacements, replacementKey, value.Value)
	}
	return strings.NewReplacer(replacements...)
}

// Replace variable references in a string
func (variables Variables) Replace(s string) string {
	replacer := variables.Replacer()
	return replacer.Replace(s)
}

// Select a subset of variables from the map
//
// Unknown keys are ignored.
func (variables Variables) Select(argList []string) Variables {
	newVariables := make(Variables)

	for _, key := range argList {
		if value, ok := variables[key]; ok {
			newVariables[key] = value
		}
	}

	return newVariables
}
