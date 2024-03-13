package validator

import (
	"encoding/json"
	"errors"
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"

	"github.com/go-playground/validator/v10"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

var oca_cacao_schemas string = "https://raw.githubusercontent.com/opencybersecurityalliance/cacao-roaster/main/lib/cacao-json-schemas/schemas/playbook.json"

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// function unmarshalls and validates againts the generic object type and return a pointer
// to the unmarshalled object.
// if input object is not in accordance with object type err is returned
func UnmarshalJson[BodyType any](b *[]byte) (any, error) {
	var body BodyType
	validate := validator.New()
	err := json.Unmarshal(*b, &body)
	if err != nil {
		return nil, err
	}

	err = validate.Struct(body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func IsValidCacaoJson(data []byte) error {

	var rawJson map[string]interface{}
	if err := json.Unmarshal([]byte(data), &rawJson); err != nil {
		return err
	}

	version := rawJson["spec_version"]

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft7

	var sch *jsonschema.Schema
	var err error
	switch version {
	case cacao.CACAO_VERSION_1:
		return errors.New("you submitted a cacap v1 playbook. at the moment, soarca only supports cacao v2 playbooks")
	case cacao.CACAO_VERSION_2:
		sch, err = compiler.Compile(oca_cacao_schemas)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported cacao version")
	}

	if err := sch.Validate(rawJson); err != nil {
		return err
	}
	return nil
}
