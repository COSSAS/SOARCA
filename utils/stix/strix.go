package stix

import (
	"errors"
	"fmt"
	"soarca/models/cacao"
	"strconv"
	"strings"
)

const (
	Equal          = "="
	NotEqual       = "!="
	Greater        = ">"
	Less           = "<"
	LessOrEqual    = "<="
	GreaterOrEqual = ">="
	In             = "IN"
	Like           = "LIKE"
	Matches        = "MATCHES"
	IsSubset       = "ISSUBSET"
	IsSuperSet     = "ISSUPERSET"
)

func Evaluate(expression string, vars cacao.Variables) (bool, error) {

	//"condition": "__variable__:value == '10.0.0.0/8'"
	//"condition": "__ip__:value/__subnet__:value == '10.0.0.0/8'"
	//"condition": "__ip__:value = __another_ip__:value"
	//"expresion_type": "ipv4"
	parts := strings.Split(expression, " ")
	if len(parts) != 3 {
		err := errors.New("comparisons can only contain 3 parts as per STIX specification")
		return false, err
	}

	usedVariable, err := findVariable(parts[0], vars)
	if err != nil {
		return false, err
	}

	parts[0] = vars.Interpolate(parts[0])

	switch usedVariable.Type {
	case cacao.VariableTypeString:
		return stringCompare(parts)
	case cacao.VariableTypeInt:
		return numberCompare(parts)
	case cacao.VariableTypeLong:
		return numberCompare(parts)
	case cacao.VariableTypeFloat:
		return floatCompare(parts)

	default:
		err := errors.New("variable type is not a cacao variable type")
		return false, err
	}

}

func findVariable(variable string, vars cacao.Variables) (cacao.Variable, error) {
	for key, value := range vars {
		replacementKey := fmt.Sprint(key, ":value")
		if strings.Contains(variable, replacementKey) {
			return value, nil
		}

	}
	return cacao.Variable{}, nil
}

func checkIpInRange(ipString string, ipRangeString string) {
	// ip, ipRange, err := net.ParseCIDR(ipString)
	// ip2, ipRange2, err2 := net.ParseCIDR(ipRangeString)
	// ipRange.Mask.Size()
}

func stringCompare(parts []string) (bool, error) {

	lhs := parts[0]
	comparator := parts[1]
	rhs := parts[2]

	switch comparator {
	case Equal:
		return strings.Compare(lhs, rhs) == 0, nil
	case NotEqual:
		return strings.Compare(lhs, rhs) != 0, nil
	case Greater:
		return strings.Compare(lhs, rhs) == 1, nil
	case Less:
		return strings.Compare(lhs, rhs) == -1, nil
	case LessOrEqual:
		return strings.Compare(lhs, rhs) <= 0, nil
	case GreaterOrEqual:
		return strings.Compare(lhs, rhs) >= 0, nil
	case In:
		return strings.Contains(lhs, rhs), nil
	default:
		err := errors.New("operator not valid")
		return false, err
	}
}

func numberCompare(parts []string) (bool, error) {
	lhs, err := strconv.Atoi(parts[0])

	if err != nil {
		return false, err
	}
	comparator := parts[1]
	rhs, err := strconv.Atoi(parts[2])
	if err != nil {
		return false, err
	}

	switch comparator {
	case Equal:
		return lhs == rhs, nil
	case NotEqual:
		return lhs != rhs, nil
	case Greater:
		return lhs > rhs, nil
	case Less:
		return lhs < rhs, nil
	case LessOrEqual:
		return lhs <= rhs, nil
	case GreaterOrEqual:
		return lhs >= rhs, nil
	default:
		err := errors.New("operator not valid")
		return false, err
	}
}

func floatCompare(parts []string) (bool, error) {
	lhs, err := strconv.ParseFloat(parts[0], 0)

	if err != nil {
		return false, err
	}
	comparator := parts[1]
	rhs, err := strconv.ParseFloat(parts[2], 0)
	if err != nil {
		return false, err
	}

	switch comparator {
	case Equal:
		return lhs == rhs, nil
	case NotEqual:
		return lhs != rhs, nil
	case Greater:
		return lhs > rhs, nil
	case Less:
		return lhs < rhs, nil
	case LessOrEqual:
		return lhs <= rhs, nil
	case GreaterOrEqual:
		return lhs >= rhs, nil
	default:
		err := errors.New("operator not valid")
		return false, err
	}
}
