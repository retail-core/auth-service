package common

import (
	"fmt"
	"net/http"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s interface{}) *AppError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errorsMap := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		field := e.Field()
		switch e.Tag() {
		case "required":
			errorsMap[field] = fmt.Sprintf("%s is required", field)
		case "email":
			errorsMap[field] = fmt.Sprintf("%s must be a valid email", field)
		case "min":
			errorsMap[field] = fmt.Sprintf("%s must be at least %s characters", field, e.Param())
		case "max":
			errorsMap[field] = fmt.Sprintf("%s must be at most %s characters", field, e.Param())
		case "oneof":
			errorsMap[field] = fmt.Sprintf("%s must be one of: %s", field, e.Param())

		default:
			errorsMap[field] = fmt.Sprintf("%s is invalid", field)
		}
	}

	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: "Invalid request parameters",
		Status:  http.StatusUnprocessableEntity,
		Details: errorsMap,
	}
}
