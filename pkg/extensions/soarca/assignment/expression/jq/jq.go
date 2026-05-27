package jq

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/extensions/soarca/assignment/expression"
	"strings"
	"time"

	"github.com/itchyny/gojq"
)

type Empty struct{}

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

const expressionEngine = "jq"

type Jq struct {
}

func New() *Jq {
	return &Jq{}
}

func (jq *Jq) GetExpressionName() string {
	return expressionEngine
}

func (jq *Jq) Execute(source string, queryExpression expression.Query) (error, string) {
	query, err := gojq.Parse(string(queryExpression))
	if err != nil {
		return err, ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	valid := json.Valid([]byte(source))
	if !valid {
		return errors.New("invalid json object"), ""
	}
	input := map[string]any{}
	err = json.Unmarshal([]byte(source), &input)
	if err != nil {
		log.Error("unmarshal failed with error: ", err.Error())
		return err, ""
	}

	it := query.RunWithContext(ctx, input)
	var result strings.Builder
	for {
		v, ok := it.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			var halt *gojq.HaltError
			if errors.As(err, &halt) && halt.Value() == nil {
				break
			}
			return err, ""
		}

		if s, ok := v.(string); ok {
			result.WriteString(s + "\n")
		} else {
			b, err := json.Marshal(v)
			if err != nil {
				return err, ""
			}
			result.Write(b)
			result.WriteByte('\n')
		}

	}

	return nil, strings.TrimSpace(result.String())
}
