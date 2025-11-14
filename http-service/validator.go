package httpservice

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()

	// Use JSON tag name instead of struct field name
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: v}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return v.formatValidationErrors(validationErrors)
		}
		return err
	}
	return nil
}

// formatValidationErrors converts validator errors to HTTPError
func (v *Validator) formatValidationErrors(errs validator.ValidationErrors) *HTTPError {
	var validationErrs []ValidationError

	for _, err := range errs {
		validationErrs = append(validationErrs, ValidationError{
			Field:   err.Field(),
			Message: v.getErrorMessage(err),
			Tag:     err.Tag(),
		})
	}

	details := make(map[string]interface{})
	details["errors"] = validationErrs

	return &HTTPError{
		Code:    422,
		Message: "Validation failed",
		Details: details,
	}
}

// getErrorMessage returns a human-readable error message
func (v *Validator) getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, err.Param())
	case "len":
		return fmt.Sprintf("%s must be %s characters long", field, err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, err.Param())
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", field, err.Param())
	case "nefield":
		return fmt.Sprintf("%s must not be equal to %s", field, err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "number":
		return fmt.Sprintf("%s must be a valid number", field)
	case "hexadecimal":
		return fmt.Sprintf("%s must be a valid hexadecimal", field)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", field)
	case "rgb":
		return fmt.Sprintf("%s must be a valid RGB color", field)
	case "rgba":
		return fmt.Sprintf("%s must be a valid RGBA color", field)
	case "hsl":
		return fmt.Sprintf("%s must be a valid HSL color", field)
	case "hsla":
		return fmt.Sprintf("%s must be a valid HSLA color", field)
	case "e164":
		return fmt.Sprintf("%s must be a valid E.164 phone number", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid3":
		return fmt.Sprintf("%s must be a valid UUID v3", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "uuid5":
		return fmt.Sprintf("%s must be a valid UUID v5", field)
	case "isbn":
		return fmt.Sprintf("%s must be a valid ISBN", field)
	case "isbn10":
		return fmt.Sprintf("%s must be a valid ISBN-10", field)
	case "isbn13":
		return fmt.Sprintf("%s must be a valid ISBN-13", field)
	case "json":
		return fmt.Sprintf("%s must be valid JSON", field)
	case "jwt":
		return fmt.Sprintf("%s must be a valid JWT", field)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude", field)
	case "ssn":
		return fmt.Sprintf("%s must be a valid SSN", field)
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", field)
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", field)
	case "ipv6":
		return fmt.Sprintf("%s must be a valid IPv6 address", field)
	case "cidr":
		return fmt.Sprintf("%s must be a valid CIDR", field)
	case "mac":
		return fmt.Sprintf("%s must be a valid MAC address", field)
	case "hostname":
		return fmt.Sprintf("%s must be a valid hostname", field)
	case "fqdn":
		return fmt.Sprintf("%s must be a valid FQDN", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime with format %s", field, err.Param())
	default:
		return fmt.Sprintf("%s failed validation for %s", field, err.Tag())
	}
}

// RegisterValidation registers a custom validation function
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// RegisterStructValidation registers a custom struct-level validation
func (v *Validator) RegisterStructValidation(fn validator.StructLevelFunc, types ...interface{}) {
	v.validate.RegisterStructValidation(fn, types...)
}
