package response

import (
	"fmt"

	"github.com/go-playground/validator"
)

type ErrorResponse struct {
	IsError bool        `json:"isError"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

func ErrorMessage(message string) ErrorResponse {
	return ErrorResponse{IsError: true, Message: message}
}

type FieldValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	ErrorResponse
	Errors []FieldValidationError `json:"errors"`
}

func ValidationError(errs validator.ValidationErrors) ValidationErrorResponse {
	var validationErrors []FieldValidationError

	for _, err := range errs {

		message := fmt.Sprintf("%s %s", err.Field(), err.Tag())

		switch err.Tag() {
		case "required":
			message = fmt.Sprintf("%s is a required field", err.Field())
		case "email":
			message = fmt.Sprintf("%s should be an valid email", err.Field())
		case "gte":
			message = fmt.Sprintf(
				"%s should be greater or equalent of %s",
				err.Field(),
				err.Param(),
			)
		}

		validationErrors = append(validationErrors, FieldValidationError{
			Field:   err.Field(),
			Message: message,
		})
	}

	return ValidationErrorResponse{
		ErrorResponse: ErrorMessage("validation failed"),
		Errors:        validationErrors,
	}
}
