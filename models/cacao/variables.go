package cacao

import (
	"fmt"
	"strings"
)

// Initialize new Variables
//
// Allows passing in multiple variables at once
func NewVariables(new ...Variable) Variables {
	variables := make(Variables)
	variables.InsertRange(new...)
	return variables
}

// Insert a variable
//
// Returns true if the variable was inserted
func (variables Variables) Insert(new Variable) bool {
	if _, found := variables[new.Name]; found {
		return false
	}
	variables[new.Name] = new
	return true
}

// Insert or replace a variable
//
// Returns true if the variable was replaced
func (variables Variables) InsertOrReplace(new Variable) bool {
	_, found := variables[new.Name]
	variables[new.Name] = new
	return found
}

// Insert multiple variables at once
func (variables Variables) InsertRange(new ...Variable) {
	for _, newVar := range new {
		variables.Insert(newVar)
	}
}

// Find a variable by name
//
// Returns a Variable struct and a boolean indicating if it was found
func (variables Variables) Find(key string) (Variable, bool) {
	val, ok := variables[key]
	return val, ok
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

// Interpolate variable references in a string
func (variables Variables) Interpolate(s string) string {
	replacer := variables.Replacer()
	return replacer.Replace(s)
}

// Select a subset of variables from the map
//
// Unknown keys are ignored.
func (variables Variables) Select(keys []string) Variables {
	newVariables := NewVariables()

	for _, key := range keys {
		if value, ok := variables.Find(key); ok {
			newVariables.InsertOrReplace(value)
		}
	}

	return newVariables
}

// Merge two maps of Cacao variables
func (variables Variables) Merge(new Variables) Variables {
	combined := NewVariables()
	for _, value := range variables {
		combined.Insert(value)
	}

	for _, newValue := range new {
		combined.InsertOrReplace(newValue)
	}

	return combined
}
