package stix

import (
	"errors"
	"soarca/models/cacao"
	"strings"
)

type StixPattern struct {
	Pattern string
}

func NewPattern(pattern string) StixPattern {
	return StixPattern{Pattern: pattern}
}

func (s StixPattern) IsTrue(variables cacao.Variables) (bool, error) {
	fields := strings.Fields(s.Pattern)
	if len(fields) != 3 {
		return false, errors.New("can only evaluate comparison expressions")
	}

	a, op, b := fields[0], fields[1], fields[2]

	if op != "=" && op != "!=" {
		return false, errors.New("can only evaluate comparison expressions")
	}

	a = variables.Interpolate(UnwrapQuotes(a))
	b = variables.Interpolate(UnwrapQuotes(b))

	if op == "=" {
		return a == b, nil
	} else {
		return a != b, nil
	}
}

func UnwrapQuotes(s string) string {
	s, _ = strings.CutPrefix(s, "'")
	s, _ = strings.CutSuffix(s, "'")
	return s
}
