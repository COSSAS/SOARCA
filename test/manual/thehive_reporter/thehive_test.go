package thehive_test

import (
	"bufio"
	"fmt"
	"os"
	"soarca/internal/reporter/downstream_reporter/thehive"
	"soarca/internal/reporter/downstream_reporter/thehive/connector"
	"strings"
	"testing"
)

// Microsoft Copilot provided code to get .env local file and extract variables values
func LoadEnv(envVar string) (string, error) {
	file, err := os.Open(".env")
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, envVar+"=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.Trim(parts[1], `"`), nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("variable %s not found", envVar)
}

func TestTheHiveConnection(t *testing.T) {
	thehive_api_tkn, err := LoadEnv("THEHIVE_TEST_API_TOKEN")
	if err != nil {
		t.Fail()
	}
	thehive_api_base_uri, err := LoadEnv("THEHIVE_TEST_API_BASE_URI")
	if err != nil {
		t.Fail()
	}
	thr := thehive.New(connector.New(thehive_api_base_uri, thehive_api_tkn))
	str := thr.ConnectorTest()
	fmt.Println(str)
}
