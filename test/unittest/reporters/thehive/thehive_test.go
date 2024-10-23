package thehive_test

import (
	"fmt"
	"soarca/internal/reporter/downstream_reporter/thehive"
	"soarca/internal/reporter/downstream_reporter/thehive/connector"
	"testing"
)

func TestTheHiveConnection(t *testing.T) {
	thr := thehive.New(connector.New("http://localhost:9000/api/v1", "JqZEDAR2vYbITI7ls45gm6WI+F8yuebZ"))
	str := thr.ConnectorTest()
	fmt.Println(str)
	t.Fail()
}
