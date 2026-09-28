package maps

// Initialize new Variables
//
// Allows passing in multiple variables at once
func New[T any]() map[string]T {
	elements := make(map[string]T)
	return elements
}

// Insert a variable
//
// Returns true if the variable was inserted
func Insert[T any](base map[string]T, key string, value T) bool {
	if _, found := base[key]; found {
		return false
	}
	base[key] = value
	return true
}

// Merge two maps of Cacao variables and replace the base with the source if exists
func Merge[T any](base map[string]T, source map[string]T) {
	for key, variable := range source {
		base[key] = variable
	}
}

// Returns true if the variable was replaced
func InsertOrReplace[T any](base map[string]T, key string, new T) bool {
	_, found := (base)[key]
	(base)[key] = new
	return found
}

// // Insert or replace a variable
// //
// // Returns true if the variable was replaced
// func (variables *Variables) InsertOrReplace(new Variable) bool {
// 	_, found := (*variables)[new.Name]
// 	(*variables)[new.Name] = new
// 	return found
// }

// // Insert variables map at once into the base and keep base variables in favor of source duplicates
// func (variables *Variables) InsertRange(source Variables) {
// 	for _, variable := range source {
// 		variables.Insert(variable)
// 	}
// }

// // Find a variable by name
// //
// // Returns a Variable struct and a boolean indicating if it was found
// func (variables Variables) Find(key string) (Variable, bool) {
// 	val, ok := variables[key]
// 	return val, ok
// }

// // Interpolate variable references into a target string
// //
// // Returns the Interpolated string with variables values available in the map
// func (variables *Variables) Interpolate(input string) string {
// 	replacements := make([]string, 0)
// 	for key, value := range *variables {
// 		replacementKey := fmt.Sprint(key, ":value")
// 		replacements = append(replacements, replacementKey, value.Value)
// 	}
// 	return strings.NewReplacer(replacements...).Replace(input)
// }

// // Select a subset of variables from the map
// //
// // Unknown keys are ignored.
// func (variables *Variables) Select(keys []string) Variables {
// 	newVariables := NewVariables()

// 	for _, key := range keys {
// 		if value, ok := variables.Find(key); ok {
// 			newVariables.InsertOrReplace(value)
// 		}
// 	}

// 	return newVariables
// }

// // Merge two maps of Cacao variables and replace the base with the source if exists
// func (variables *Variables) Merge(source Variables) {
// 	for _, variable := range source {
// 		variables.InsertOrReplace(variable)
// 	}
// }
