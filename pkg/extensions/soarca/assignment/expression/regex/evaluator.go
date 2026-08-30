package regex

import (
	"reflect"
	"regexp"
	"soarca/internal/logger"
	"soarca/pkg/extensions/soarca/assignment/expression"
	"strings"
)

type Empty struct{}

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

const expressionEngine = "regex"

type Regex struct {
}

func New() *Regex {
	return &Regex{}
}

func (regex *Regex) GetEngineName() string {
	return expressionEngine
}

// Execute compiles queryExpression as a Go regular expression and applies it to
// source. The result depends on the pattern:
//
//   - With one or more capture groups, the first capture group of every match
//     is returned, one per line.
//   - Without capture groups, the full match for every occurrence is returned,
//     one per line.
//
// A pattern that produces no matches returns an empty string with no error; a
// pattern that fails to compile returns the compile error.
func (regex *Regex) Execute(source string, queryExpression expression.Query) (string, error) {
	pattern, err := regexp.Compile(string(queryExpression))
	if err != nil {
		log.Error("regex compile failed with error: ", err.Error())
		return "", err
	}

	matches := pattern.FindAllStringSubmatch(source, -1)
	if len(matches) == 0 {
		return "", nil
	}

	hasCaptureGroup := pattern.NumSubexp() > 0
	var result strings.Builder
	for _, match := range matches {
		if hasCaptureGroup {
			result.WriteString(match[1])
		} else {
			result.WriteString(match[0])
		}
		result.WriteByte('\n')
	}

	return strings.TrimSpace(result.String()), nil
}
