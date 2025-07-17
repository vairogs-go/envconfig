package envconfig

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidationError represents a single validation error for a specific field.
type ValidationError struct {
	Field   string // The field name that failed validation
	Message string // The validation error message
}

// Error implements the error interface for ValidationError.
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors, combining all
// individual validation errors into a single error message.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	messages := make([]string, 0, len(e))
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("validation failed with %d error(s): %s", len(e), strings.Join(messages, "; "))
}

// Validator defines the interface for struct validation.
type Validator interface {
	ValidateStruct(cfg any) error
}

// StructValidator implements the Validator interface for validating structs
// using reflection and struct tags.
type StructValidator struct{}

// NewValidator creates a new instance of StructValidator.
func NewValidator() *StructValidator {
	return &StructValidator{}
}

// ValidateStruct validates a struct using reflection and struct tags.
// It supports the following validation tags:
// - required: "true" - field must not be empty
// - min: "N" - minimum length for strings
// - max: "N" - maximum length for strings
// - pattern: "alphanumeric" - field must match the specified pattern
//
// The function recursively validates nested structs and handles pointers.
//
// Example:
//
//	type Config struct {
//	  Host string `required:"true" min:"1" max:"255"`
//	  Port int    `min:"1" max:"65535"`
//	}
//
//	validator := NewValidator()
//	err := validator.ValidateStruct(&Config{})
func (v *StructValidator) ValidateStruct(cfg any) error {
	if cfg == nil {
		return ValidationError{Field: "config", Message: "configuration cannot be nil"}
	}

	reflectValue := reflect.ValueOf(cfg)
	if reflectValue.Kind() == reflect.Ptr {
		if reflectValue.IsNil() {
			return ValidationError{Field: "config", Message: "configuration pointer cannot be nil"}
		}
		reflectValue = reflectValue.Elem()
	}

	if reflectValue.Kind() != reflect.Struct {
		return ValidationError{Field: "config", Message: "configuration must be a struct"}
	}

	var errors ValidationErrors
	v.validateStructInner(reflectValue, "", &errors)

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateStructInner recursively validates struct fields and nested structs.
func (v *StructValidator) validateStructInner(value reflect.Value, prefix string, errors *ValidationErrors) {
	if value.Kind() != reflect.Struct {
		return
	}

	structType := value.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := value.Field(i)
		fieldName := v.getFieldName(field, prefix)

		if !field.IsExported() {
			continue
		}

		v.validateField(field, fieldValue, fieldName, errors)
		if fieldValue.Kind() == reflect.Struct {
			v.validateStructInner(fieldValue, fieldName, errors)
		} else if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() && fieldValue.Elem().Kind() == reflect.Struct {
			v.validateStructInner(fieldValue.Elem(), fieldName, errors)
		}
	}
}

// validateField validates a single field according to its struct tags.
func (v *StructValidator) validateField(field reflect.StructField, value reflect.Value,
	fieldName string, errors *ValidationErrors) {
	if v.isRequired(field) && v.isEmpty(value) {
		*errors = append(*errors, ValidationError{
			Field:   fieldName,
			Message: "required field is empty",
		})
		return
	}

	if value.Kind() == reflect.String {
		v.validateStringField(field, value, fieldName, errors)
	}
}

// validateStringField validates string fields for length and pattern constraints.
func (v *StructValidator) validateStringField(field reflect.StructField, value reflect.Value,
	fieldName string, errors *ValidationErrors) {
	str := value.String()

	if minLen := field.Tag.Get("min"); minLen != "" {
		if len(str) < v.parseInt(minLen, 0) {
			*errors = append(*errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("minimum length is %s characters", minLen),
			})
		}
	}

	if maxLen := field.Tag.Get("max"); maxLen != "" {
		if len(str) > v.parseInt(maxLen, 0) {
			*errors = append(*errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("maximum length is %s characters", maxLen),
			})
		}
	}

	if pattern := field.Tag.Get("pattern"); pattern != "" && str != "" {
		if !v.matchesPattern(str, pattern) {
			*errors = append(*errors, ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("does not match required pattern: %s", pattern),
			})
		}
	}
}

// getFieldName extracts the field name for validation error messages.
func (v *StructValidator) getFieldName(field reflect.StructField, prefix string) string {
	name := field.Tag.Get("mapstructure")
	if name == "" {
		name = strings.ToLower(field.Name)
	}
	if prefix != "" {
		return prefix + "." + name
	}
	return name
}

// isRequired checks if a field has the required tag set to "true".
func (v *StructValidator) isRequired(field reflect.StructField) bool {
	return field.Tag.Get("required") == "true"
}

// isEmpty checks if reflect.Value is considered empty based on its type.
func (v *StructValidator) isEmpty(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.String() == ""
	case reflect.Ptr:
		return value.IsNil()
	case reflect.Slice, reflect.Array, reflect.Map:
		return value.Len() == 0
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Chan,
		reflect.Func, reflect.Interface, reflect.Struct, reflect.UnsafePointer:
		return false
	default:
		return false
	}
}

// parseInt parses a string to an integer, returning a default value on error.
func (v *StructValidator) parseInt(s string, defaultVal int) int {
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return defaultVal
	}
	return result
}

// matchesPattern checks if a string matches a predefined pattern.
func (v *StructValidator) matchesPattern(s, pattern string) bool {
	switch pattern {
	case "alphanumeric":
		return v.isAlphanumeric(s)
	default:
		return true
	}
}

// isAlphanumeric checks if a string contains only alphanumeric characters.
func (v *StructValidator) isAlphanumeric(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}
