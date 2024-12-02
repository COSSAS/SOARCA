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
	for _, variable := range new {
		variables.Insert(variable)
	}
	return variables
}

// Insert a variable
//
// Returns true if the variable was inserted
func (variables *Variables) Insert(new Variable) bool {
	if _, found := (*variables)[new.Name]; found {
		return false
	}
	(*variables)[new.Name] = new
	return true
}

// Insert or replace a variable
//
// Returns true if the variable was replaced
func (variables *Variables) InsertOrReplace(new Variable) bool {
	_, found := (*variables)[new.Name]
	(*variables)[new.Name] = new
	return found
}

// Insert variables map at once into the base and keep base variables in favor of source duplicates
func (variables *Variables) InsertRange(source Variables) {
	for _, variable := range source {
		variables.Insert(variable)
	}
}

// Find a variable by name
//
// Returns a Variable struct and a boolean indicating if it was found
func (variables Variables) Find(key string) (Variable, bool) {
	val, ok := variables[key]
	return val, ok
}

// Interpolate variable references into a target string
//
// Returns the Interpolated string with variables values available in the map
func (variables *Variables) Interpolate(input string) string {
	replacements := make([]string, 0)
	for key, value := range *variables {
		replacementKey := fmt.Sprint(key, ":value")
		replacements = append(replacements, replacementKey, value.Value)
	}
	return strings.NewReplacer(replacements...).Replace(input)
}

// Select a subset of variables from the map
//
// Unknown keys are ignored.
func (variables *Variables) Select(keys []string) Variables {
	newVariables := NewVariables()

	for _, key := range keys {
		if value, ok := variables.Find(key); ok {
			newVariables.InsertOrReplace(value)
		}
	}

	return newVariables
}

// Merge two maps of Cacao variables and replace the base with the source if exists
func (variables *Variables) Merge(source Variables) {
	for _, variable := range source {
		variables.InsertOrReplace(variable)
	}
}
