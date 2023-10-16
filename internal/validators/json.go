package validation

import (
	"encoding/json"

	validator "github.com/go-playground/validator/v10"
)

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
