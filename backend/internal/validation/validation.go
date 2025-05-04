package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register function to get json tag names instead of struct field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func Validate(s interface{}) error {
	return validate.Struct(s)
}

func ValidationErrors(err error) map[string]string {
	if err == nil {
		return nil
	}

	errs := make(map[string]string)

	validationErrors := err.(validator.ValidationErrors)
	for _, e := range validationErrors {
		errs[e.Field()] = validationErrorMsg(e)
	}

	return errs
}

func validationErrorMsg(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value must be at least " + e.Param()
	case "max":
		return "Value cannot be longer than " + e.Param()
	default:
		return "Invalid value"
	}
}
