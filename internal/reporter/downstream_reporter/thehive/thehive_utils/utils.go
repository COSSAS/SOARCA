package thehive_utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"soarca/logger"
	"strings"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// ############################### Utils

func CleanUrlString(url string) string {
	// Replace double slashes in the URL after http(s)://
	parts := strings.SplitN(url, "//", 2)
	cleanedPath := strings.ReplaceAll(parts[1], "//", "/")
	url = parts[0] + "//" + cleanedPath
	return url
}

func MarhsalRequestBody(body interface{}) (io.Reader, error) {
	var requestBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling JSON: %v", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}
	log.Debug(fmt.Sprintf("request body: %s", requestBody))
	return requestBody, nil
}

func GetIdFromArrayBody(body []byte) (string, error) {
	// Try to unmarshal as a slice of maps
	var respArray []map[string]interface{}
	err := json.Unmarshal(body, &respArray)
	if err != nil {
		return "", err
	}

	if len(respArray) == 0 {
		return "", errors.New("response array is empty")
	}

	_id, ok := respArray[0]["_id"].(string)
	if !ok {
		return "", errors.New("type assertion for retrieving TheHive ID failed")
	}
	return _id, nil
}

func GetIdFromObjectBody(body []byte) (string, error) {
	// If unmarshalling as a slice fails, try to unmarshal as a single map
	var respMap map[string]interface{}
	err := json.Unmarshal(body, &respMap)
	if err != nil {
		return "", err
	}

	_id, ok := respMap["_id"].(string)
	if !ok {
		return "", errors.New("type assertion for retrieving TheHive ID from map failed")
	}

	return _id, nil
}

func FormatVariable(s interface{}) string {
	v := reflect.ValueOf(s)
	t := v.Type()

	var sb strings.Builder

	sb.WriteString(strings.Repeat("*", 30) + "\n")

	maxFieldLength := 0
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i).Name
		if len(field) > maxFieldLength {
			maxFieldLength = len(field)
		}
	}

	var name, value, description string
	var otherFields []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i).Name
		fieldValue := fmt.Sprintf("%v", v.Field(i).Interface())
		padding := strings.Repeat("\t", (maxFieldLength-len(field))/4+1)
		switch field {
		case "Name":
			name = fmt.Sprintf("%s:%s%s\n", field, padding, fieldValue)
		case "Value":
			value = fmt.Sprintf("%s:%s%s\n", field, padding, fieldValue)
		case "Description":
			description = fmt.Sprintf("%s:%s%s\n", field, padding, fieldValue)
		default:
			otherFields = append(otherFields, fmt.Sprintf("%s:%s%s\n", field, padding, fieldValue))
		}
	}

	// Append "Name", "Description", and "Value" fields first
	sb.WriteString(name)
	sb.WriteString(description)
	sb.WriteString(value)

	// Append all other fields
	for _, field := range otherFields {
		sb.WriteString(field)
	}

	sb.WriteString(strings.Repeat("*", 30) + "\n")

	return sb.String()
}
