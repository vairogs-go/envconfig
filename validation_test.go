package envconfig

import (
	"reflect"
	"strings"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "validation error for field 'test_field': test message"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   ValidationErrors
		expected string
	}{
		{
			name:     "empty errors",
			errors:   ValidationErrors{},
			expected: "no validation errors",
		},
		{
			name: "single error",
			errors: ValidationErrors{
				{Field: "field1", Message: "error1"},
			},
			expected: "validation failed with 1 error(s): validation error for field 'field1': error1",
		},
		{
			name: "multiple errors",
			errors: ValidationErrors{
				{Field: "field1", Message: "error1"},
				{Field: "field2", Message: "error2"},
			},
			expected: "validation failed with 2 error(s): validation error for field 'field1': error1; validation error for field 'field2': error2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Error()
			if result != tt.expected {
				t.Errorf("ValidationErrors.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Error("NewValidator() returned nil")
	}

	// Check that it implements the Validator interface
	var _ Validator = validator
}

func TestStructValidator_ValidateStruct(t *testing.T) {
	validator := NewValidator()

	type TestStruct struct {
		RequiredField string `required:"true"`
		OptionalField string
		MinField      string `min:"3"`
		MaxField      string `max:"10"`
		PatternField  string `pattern:"alphanumeric"`
		ComboField    string `required:"true" min:"2" max:"5"`
	}

	tests := []struct {
		name      string
		config    any
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil config",
			config:    nil,
			wantError: true,
			errorMsg:  "configuration cannot be nil",
		},
		{
			name:      "nil pointer",
			config:    (*TestStruct)(nil),
			wantError: true,
			errorMsg:  "configuration pointer cannot be nil",
		},
		{
			name:      "non-struct",
			config:    "not a struct",
			wantError: true,
			errorMsg:  "configuration must be a struct",
		},
		{
			name: "valid struct",
			config: &TestStruct{
				RequiredField: "required",
				OptionalField: "optional",
				MinField:      "min",
				MaxField:      "max",
				PatternField:  "abc123",
				ComboField:    "combo",
			},
			wantError: false,
		},
		{
			name: "missing required field",
			config: &TestStruct{
				OptionalField: "optional",
			},
			wantError: true,
			errorMsg:  "requiredfield",
		},
		{
			name: "field too short",
			config: &TestStruct{
				RequiredField: "required",
				MinField:      "ab",
			},
			wantError: true,
			errorMsg:  "minimum length",
		},
		{
			name: "field too long",
			config: &TestStruct{
				RequiredField: "required",
				MaxField:      "this is way too long",
			},
			wantError: true,
			errorMsg:  "maximum length",
		},
		{
			name: "pattern mismatch",
			config: &TestStruct{
				RequiredField: "required",
				PatternField:  "abc@123",
			},
			wantError: true,
			errorMsg:  "does not match required pattern",
		},
		{
			name: "combo field validation",
			config: &TestStruct{
				RequiredField: "required",
				ComboField:    "a", // Too short
			},
			wantError: true,
			errorMsg:  "minimum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStruct(tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("ValidateStruct() expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateStruct() error = %v, want to contain %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateStruct() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestStructValidator_ValidateStructWithNestedStruct(t *testing.T) {
	validator := NewValidator()

	type NestedStruct struct {
		NestedRequired string `required:"true"`
		NestedOptional string
	}

	type TestStruct struct {
		RequiredField string `required:"true"`
		Nested        NestedStruct
	}

	tests := []struct {
		name      string
		config    TestStruct
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid nested struct",
			config: TestStruct{
				RequiredField: "required",
				Nested: NestedStruct{
					NestedRequired: "nested_required",
					NestedOptional: "nested_optional",
				},
			},
			wantError: false,
		},
		{
			name: "missing nested required field",
			config: TestStruct{
				RequiredField: "required",
				Nested: NestedStruct{
					NestedOptional: "nested_optional",
				},
			},
			wantError: true,
			errorMsg:  "nested.nestedrequired",
		},
		{
			name: "missing top-level required field",
			config: TestStruct{
				Nested: NestedStruct{
					NestedRequired: "nested_required",
				},
			},
			wantError: true,
			errorMsg:  "requiredfield",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStruct(&tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("ValidateStruct() expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateStruct() error = %v, want to contain %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateStruct() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestStructValidator_ValidateStructWithPointerField(t *testing.T) {
	validator := NewValidator()

	type NestedStruct struct {
		NestedRequired string `required:"true"`
	}

	type TestStruct struct {
		RequiredField string `required:"true"`
		NestedPtr     *NestedStruct
	}

	tests := []struct {
		name      string
		config    TestStruct
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid pointer field",
			config: TestStruct{
				RequiredField: "required",
				NestedPtr: &NestedStruct{
					NestedRequired: "nested_required",
				},
			},
			wantError: false,
		},
		{
			name: "nil pointer field",
			config: TestStruct{
				RequiredField: "required",
				NestedPtr:     nil,
			},
			wantError: false, // nil pointers are not validated
		},
		{
			name: "invalid nested field in pointer",
			config: TestStruct{
				RequiredField: "required",
				NestedPtr:     &NestedStruct{
					// Missing NestedRequired
				},
			},
			wantError: true,
			errorMsg:  "nestedptr.nestedrequired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStruct(&tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("ValidateStruct() expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("ValidateStruct() error = %v, want to contain %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateStruct() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestStructValidator_getFieldName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		field    reflect.StructField
		prefix   string
		expected string
	}{
		{
			name:     "field with mapstructure tag",
			field:    reflect.StructField{Name: "TestField", Tag: `mapstructure:"custom_name"`},
			prefix:   "",
			expected: "custom_name",
		},
		{
			name:     "field without mapstructure tag",
			field:    reflect.StructField{Name: "TestField"},
			prefix:   "",
			expected: "testfield",
		},
		{
			name:     "field with prefix",
			field:    reflect.StructField{Name: "TestField"},
			prefix:   "parent",
			expected: "parent.testfield",
		},
		{
			name:     "field with mapstructure tag and prefix",
			field:    reflect.StructField{Name: "TestField", Tag: `mapstructure:"custom_name"`},
			prefix:   "parent",
			expected: "parent.custom_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.getFieldName(tt.field, tt.prefix)
			if result != tt.expected {
				t.Errorf("getFieldName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStructValidator_isRequired(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		field    reflect.StructField
		expected bool
	}{
		{
			name:     "required field",
			field:    reflect.StructField{Tag: `required:"true"`},
			expected: true,
		},
		{
			name:     "not required field",
			field:    reflect.StructField{Tag: `required:"false"`},
			expected: false,
		},
		{
			name:     "field without required tag",
			field:    reflect.StructField{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isRequired(tt.field)
			if result != tt.expected {
				t.Errorf("isRequired() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStructValidator_isEmpty(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		value    reflect.Value
		expected bool
	}{
		{
			name:     "empty string",
			value:    reflect.ValueOf(""),
			expected: true,
		},
		{
			name:     "non-empty string",
			value:    reflect.ValueOf("hello"),
			expected: false,
		},
		{
			name:     "nil pointer",
			value:    reflect.ValueOf((*string)(nil)),
			expected: true,
		},
		{
			name:     "non-nil pointer",
			value:    func() reflect.Value { s := "test"; return reflect.ValueOf(&s) }(),
			expected: false,
		},
		{
			name:     "empty slice",
			value:    reflect.ValueOf([]string{}),
			expected: true,
		},
		{
			name:     "non-empty slice",
			value:    reflect.ValueOf([]string{"item"}),
			expected: false,
		},
		{
			name:     "empty array",
			value:    reflect.ValueOf([0]string{}),
			expected: true,
		},
		{
			name:     "non-empty array",
			value:    reflect.ValueOf([1]string{"item"}),
			expected: false,
		},
		{
			name:     "empty map",
			value:    reflect.ValueOf(map[string]string{}),
			expected: true,
		},
		{
			name:     "non-empty map",
			value:    reflect.ValueOf(map[string]string{"key": "value"}),
			expected: false,
		},
		{
			name:     "struct (never empty)",
			value:    reflect.ValueOf(struct{}{}),
			expected: false,
		},
		{
			name:     "int (never empty)",
			value:    reflect.ValueOf(0),
			expected: false,
		},
		{
			name:     "bool (never empty)",
			value:    reflect.ValueOf(false),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isEmpty(tt.value)
			if result != tt.expected {
				t.Errorf("isEmpty() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStructValidator_parseInt(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name       string
		input      string
		defaultVal int
		expected   int
	}{
		{
			name:       "valid integer",
			input:      "123",
			defaultVal: 0,
			expected:   123,
		},
		{
			name:       "negative integer",
			input:      "-456",
			defaultVal: 0,
			expected:   -456,
		},
		{
			name:       "invalid string",
			input:      "abc",
			defaultVal: 10,
			expected:   10,
		},
		{
			name:       "empty string",
			input:      "",
			defaultVal: 5,
			expected:   5,
		},
		{
			name:       "float string",
			input:      "123.45",
			defaultVal: 0,
			expected:   123, // parseInt reads up to the first non-digit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.parseInt(tt.input, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("parseInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStructValidator_matchesPattern(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		pattern  string
		expected bool
	}{
		{
			name:     "alphanumeric valid",
			input:    "abc123",
			pattern:  "alphanumeric",
			expected: true,
		},
		{
			name:     "alphanumeric invalid",
			input:    "abc@123",
			pattern:  "alphanumeric",
			expected: false,
		},
		{
			name:     "alphanumeric empty",
			input:    "",
			pattern:  "alphanumeric",
			expected: true,
		},
		{
			name:     "unknown pattern",
			input:    "anything",
			pattern:  "unknown",
			expected: true, // Unknown patterns return true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.matchesPattern(tt.input, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesPattern() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStructValidator_isAlphanumeric(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid alphanumeric",
			input:    "abc123",
			expected: true,
		},
		{
			name:     "valid letters only",
			input:    "abcDEF",
			expected: true,
		},
		{
			name:     "valid numbers only",
			input:    "123456",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "with special characters",
			input:    "abc@123",
			expected: false,
		},
		{
			name:     "with spaces",
			input:    "abc 123",
			expected: false,
		},
		{
			name:     "with underscore",
			input:    "abc_123",
			expected: false,
		},
		{
			name:     "with dash",
			input:    "abc-123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isAlphanumeric(tt.input)
			if result != tt.expected {
				t.Errorf("isAlphanumeric() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStructValidator_validateStructWithUnexportedFields(t *testing.T) {
	validator := NewValidator()

	type TestStruct struct {
		ExportedField   string `required:"true"`
		unexportedField string `required:"true"` // This should be ignored
	}

	// Only exported fields should be validated
	config := TestStruct{
		// Missing ExportedField (should cause error)
		unexportedField: "", // This should be ignored
	}

	err := validator.ValidateStruct(&config)
	if err == nil {
		t.Error("ValidateStruct() expected error for missing exported field")
		return
	}

	// Should only complain about the exported field
	if !strings.Contains(err.Error(), "exportedfield") {
		t.Errorf("ValidateStruct() error = %v, want to contain 'exportedfield'", err.Error())
	}

	// Should not complain about unexported field
	if strings.Contains(err.Error(), "unexportedfield") {
		t.Errorf("ValidateStruct() error = %v, should not contain 'unexportedfield'", err.Error())
	}
}
