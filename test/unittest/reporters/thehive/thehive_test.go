package thehive_test

import (
	"fmt"
	"soarca/internal/reporter/downstream_reporter/thehive"
	"soarca/internal/reporter/downstream_reporter/thehive/connector"
	"soarca/utils"
	"testing"
)

func TestTheHiveConnection(t *testing.T) {
	thr := thehive.New(connector.New(utils.GetEnv("THEHIVE_TEST_API_TOKEN", ""), utils.GetEnv("THEHIVE_TEST_API_BASE_URI", "")))
	str := thr.ConnectorTest()
	fmt.Println(str)
	t.Fail()
}
