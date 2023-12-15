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

var cacao_v1_csd01_http string = "https://raw.githubusercontent.com/cyentific-rni/cacao-json-schemas/cacao-v1.0-csd02/schemas/playbook.json"
var cacao_v2_csd01_http string = "https://raw.githubusercontent.com/cyentific-rni/cacao-json-schemas/cacao-v2.0-csd01/schemas/playbook.json"

//var cacao_v2_csd03_http string = "https://raw.githubusercontent.com/cyentific-rni/cacao-json-schemas/cacao-v2.0-csd03/schemas/playbook.json"

func init() {
	log = logger.Logger(component, logger.Trace, "", logger.Json)
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

// TODO: change return to just error as boolean does not provide additional info

func IsValidCacaoJson(data []byte) error {
	// NOTE: Using cacao-v2.0-csd01 instead of cacao-v2.0-csd03
	// Because latest version seem to have a bug. Opened issue in repo

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
		sch, err = compiler.Compile(cacao_v1_csd01_http)
		if err != nil {
			return err
		}
	case cacao.CACAO_VERSION_2:
		// NOTE: CURRENTLY THERE IS AN INCONSISTENCY BETWEEN CDS01 AND CDS03
		// The cds03 schema is bugged at the time being (13/11/2023)
		// So we cannot validate checking authentication information
		sch, err = compiler.Compile(cacao_v2_csd01_http)
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
