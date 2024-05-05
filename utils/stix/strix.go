package stix

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
	"strconv"
	"strings"

	"github.com/google/uuid"
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

type Empty struct{}

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

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
	case cacao.VariableTypeBool:
		return boolCompare(parts)
	case cacao.VariableTypeString:
		return stringCompare(parts)
	case cacao.VariableTypeInt:
		return numberCompare(parts)
	case cacao.VariableTypeLong:
		return numberCompare(parts)
	case cacao.VariableTypeFloat:
		return floatCompare(parts)
	case cacao.VariableTypeIpv4Address:
		return compareIp(parts)
	case cacao.VariableTypeIpv6Address:
		return compareIp(parts)
	case cacao.VariableTypeMacAddress:
		return compareMac(parts)
	case cacao.VariableTypeHash:
		// Just compare the hash strings
		return compareHash(parts)
	case cacao.VariableTypeMd5Has:
		// Just compare the hash strings
		return compareHash(parts)
	case cacao.VariableTypeSha256:
		// Just compare the hash strings
		return compareHash(parts)
	case cacao.VariableTypeHexString:
		return stringCompare(parts)
	case cacao.VariableTypeUri:
		return compareUri(parts)
	case cacao.VariableTypeUuid:
		return compareUuid(parts)
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
		err := errors.New("operator: " + comparator + " not valid or implemented")
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
		err := errors.New("operator: " + comparator + " not valid or implemented")
		return false, err
	}
}

func floatCompare(parts []string) (bool, error) {
	lhs, err := strconv.ParseFloat(parts[0], 64)

	if err != nil {
		return false, err
	}
	comparator := parts[1]
	rhs, err := strconv.ParseFloat(parts[2], 64)
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
		err := errors.New("operator: " + comparator + " not valid or implemented")
		return false, err
	}
}

func boolCompare(parts []string) (bool, error) {
	lhs, err := strconv.ParseBool(parts[0])
	if err != nil {
		return false, err
	}
	comparator := parts[1]
	rhs, err := strconv.ParseBool(parts[2])
	if err != nil {
		return false, err
	}
	switch comparator {
	case Equal:
		return lhs == rhs, nil
	case NotEqual:
		return lhs != rhs, nil
	default:
		err := errors.New("operator: " + comparator + " not valid or implemented")
		return false, err
	}
}

func compareIp(parts []string) (bool, error) {
	lhsIp := net.ParseIP(parts[0])

	comparator := parts[1]
	rhsIp := net.ParseIP(parts[2])
	switch comparator {
	case Equal:
		return lhsIp.Equal(rhsIp), nil
	case NotEqual:
		return !lhsIp.Equal(rhsIp), nil
	case In:
		_, rhsNet, err := net.ParseCIDR(parts[2])
		if err != nil {
			return false, err
		}
		return rhsNet.Contains(lhsIp), err
	default:
		err := errors.New("operator: " + comparator + " not valid or implemented")
		return false, err
	}
}

// Validate if they are hardware MAC addresses and do a string compare
func compareMac(parts []string) (bool, error) {
	lhs, err := net.ParseMAC(parts[0])
	if err != nil {
		return false, err
	}

	rhs, err := net.ParseMAC(parts[2])
	if err != nil {
		return false, err
	}

	newParts := []string{lhs.String(), parts[1], rhs.String()}

	return stringCompare(newParts)
}

func compareHash(parts []string) (bool, error) {
	lhs := parts[0]
	rhs := parts[2]
	if len(lhs) != len(rhs) {
		log.Warning("hash lengths do not match")
	}

	switch len(lhs) {
	case 32:
		log.Trace("MD5 type hash")
	case 40:
		log.Trace("SHA1 type hash")
	case 56:
		log.Trace("SHA224 type hash")
	case 64:
		log.Trace("SHA256 type hash")
	case 96:
		log.Trace("SHA384 type hash")
	case 128:
		log.Trace("SHA512 type hash")
	default:
		log.Warning("unknown hash length of: " + string(len(lhs)))
	}

	return stringCompare(parts)
}

func compareUri(parts []string) (bool, error) {
	lhs, err := url.Parse(parts[0])
	if err != nil {
		return false, err
	}
	comparator := parts[1]
	rhs, err := url.Parse(parts[2])
	if err != nil {
		return false, err
	}

	switch comparator {
	case Equal:
		return lhs.String() == rhs.String(), nil
	case NotEqual:
		return lhs.String() != rhs.String(), nil
	default:
		err := errors.New("operator: " + comparator + " not valid or implemented")
		return false, err
	}
}

func compareUuid(parts []string) (bool, error) {
	lhs, err := uuid.Parse(parts[0])
	if err != nil {
		return false, err
	}
	comparator := parts[1]
	rhs, err := uuid.Parse(parts[2])
	if err != nil {
		return false, err
	}

	switch comparator {
	case Equal:
		return lhs == rhs, nil
	case NotEqual:
		return lhs != rhs, nil
	default:
		err := errors.New("operator: " + comparator + " not valid or implemented")
		return false, err
	}
}
