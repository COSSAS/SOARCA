package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"soarca/internal/logger"
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

func StructToMDJSON(v interface{}) (string, error) {

	indentedJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling to JSON: %w", err)
	}

	markdownContent := "```json\n" + string(indentedJSON) + "\n```"
	return markdownContent, nil
}

func StructToTxtJSON(v interface{}) (string, error) {

	indentedJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling to JSON: %w", err)
	}

	txtContent := string(indentedJSON) + "\n"
	return txtContent, nil
}
