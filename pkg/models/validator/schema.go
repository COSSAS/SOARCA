package validator

import (
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/utils"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

const (
	oca_cacao_schemas string = "./schemas/playbook.json"
)

//go:embed schemas/*
var schemas embed.FS

//var oasis_cacao_schemas string = "https://raw.githubusercontent.com/oasis-open/cacao-json-schemas/main/schemas/playbook.json"
//var cyentific_cacao_schemas string = "https://raw.githubusercontent.com/cyentific-rni/cacao-json-schemas/cacao-v2.0-csd03/schemas/playbook.json"

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

func validateWithLocalSchema(playbookToValidate map[string]interface{}) error {

	compiler := jsonschema.NewCompiler()

	err := fs.WalkDir(schemas, ".", func(path string, d fs.DirEntry, err error) error {
		isFile := d.Type().IsRegular()

		if isFile {
			content, _ := fs.ReadFile(schemas, path)
			data, err := jsonschema.UnmarshalJSON(strings.NewReader(string(content)))
			if err != nil {
				return err
			}
			if err := compiler.AddResource(path, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error(err)
		return err
	}

	sch, err := compiler.Compile(oca_cacao_schemas)
	if err != nil {
		return err
	}

	err = sch.Validate(playbookToValidate)
	return err
}

func validateWithRemoteSchema(data map[string]interface{}, url string) error {
	compiler := jsonschema.NewCompiler()

	sch, err := compiler.Compile(url)
	if err != nil {
		return err
	}
	if err := sch.Validate(data); err != nil {
		return err
	}
	return nil
}

func IsValidCacaoJson(data []byte) error {

	var rawJson map[string]interface{}
	if err := json.Unmarshal([]byte(data), &rawJson); err != nil {
		return err
	}

	version := rawJson["spec_version"]

	switch version {
	case cacao.CACAO_VERSION_1:
		return errors.New("you submitted a cacao v1 playbook. at the moment, soarca only supports cacao v2 playbooks")
	case cacao.CACAO_VERSION_2:
		schemaUrl := utils.GetEnv("VALIDATION_SCHEMA_URL", "")
		if schemaUrl != "" {
			return validateWithRemoteSchema(rawJson, schemaUrl)
		} else {
			return validateWithLocalSchema(rawJson)
		}
	default:
		return errors.New("unsupported cacao version")
	}

}
