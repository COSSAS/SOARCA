package validation

import (
	"encoding/json"
	"fmt"

	validator "github.com/go-playground/validator/v10"
)

func Json[BodyType any](b []byte) error {
	var body BodyType
	validate := validator.New()
	
    if err := json.Unmarshal(b, &body); err != nil { 
        return err
    }
	if err := validate.Struct(body); err != nil { 
		fmt.Println(err.Error()) // needs to be changed with logging
        return err
    }
	return nil
}


