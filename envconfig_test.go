package envconfig

import (
	"reflect"
	"testing"
)

func TestReadDotenvBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected map[string]string
	}{
		{
			name:     "empty input",
			input:    []byte(""),
			expected: map[string]string{},
		},
		{
			name:     "simple key-value pairs",
			input:    []byte("KEY1=value1\nKEY2=value2"),
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "with comments and empty lines",
			input:    []byte("# Comment\nKEY1=value1\n\nKEY2=value2\n# Another comment"),
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "with whitespace",
			input:    []byte("  KEY1  =  value1  \n\t KEY2 \t= \t value2 \t"),
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
		{
			name:     "empty values",
			input:    []byte("KEY1=\nKEY2=value2"),
			expected: map[string]string{"KEY1": "", "KEY2": "value2"},
		},
		{
			name:     "values with equals signs",
			input:    []byte("KEY1=value=with=equals\nKEY2=value2"),
			expected: map[string]string{"KEY1": "value=with=equals", "KEY2": "value2"},
		},
		{
			name:     "malformed lines without equals",
			input:    []byte("KEY1=value1\nINVALID_LINE\nKEY2=value2"),
			expected: map[string]string{"KEY1": "value1", "KEY2": "value2"},
		},
	}

	for _, testT := range tests {
		t.Run(testT.name, func(t *testing.T) {
			result := ReadDotenvBytes(testT.input)
			if !reflect.DeepEqual(result, testT.expected) {
				t.Errorf("ReadDotenvBytes() = %v, want %v", result, testT.expected)
			}
		})
	}
}

func TestGetEnvs(t *testing.T) {
	// Set a test environment variable
	t.Setenv("TEST_ENV_VAR", "test_value")

	result := GetEnvs()

	// Check that the result is a map
	if result == nil {
		t.Fatal("GetEnvs() returned nil")
	}

	// Check that our test environment variable is included
	if value, exists := result["TEST_ENV_VAR"]; !exists || value != "test_value" {
		t.Errorf("Expected TEST_ENV_VAR=test_value, got %v=%v", "TEST_ENV_VAR", value)
	}

	// Check that common environment variables exist (these should be present in most systems)
	if len(result) == 0 {
		t.Error("GetEnvs() returned empty map, expected some environment variables")
	}
}

func TestMergeEnvMaps(t *testing.T) {
	tests := []struct {
		name     string
		input    []map[string]string
		expected map[string]string
	}{
		{
			name:     "empty maps",
			input:    []map[string]string{},
			expected: map[string]string{},
		},
		{
			name:     "single map",
			input:    []map[string]string{{"KEY1": "value1", "KEY2": "value2"}},
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "multiple maps with override",
			input: []map[string]string{
				{"KEY1": "value1", "KEY2": "value2"},
				{"KEY1": "override1", "KEY3": "value3"},
			},
			expected: map[string]string{"key1": "override1", "key2": "value2", "key3": "value3"},
		},
		{
			name: "underscore to dot conversion",
			input: []map[string]string{
				{"DATABASE_HOST": "localhost", "DATABASE_PORT": "5432"},
			},
			expected: map[string]string{"database.host": "localhost", "database.port": "5432"},
		},
		{
			name: "empty values are skipped",
			input: []map[string]string{
				{"KEY1": "value1", "KEY2": ""},
				{"KEY3": "value3"},
			},
			expected: map[string]string{"key1": "value1", "key3": "value3"},
		},
		{
			name: "case normalization",
			input: []map[string]string{
				{"Key1": "value1", "KEY2": "value2", "key3": "value3"},
			},
			expected: map[string]string{"key1": "value1", "key2": "value2", "key3": "value3"},
		},
	}

	for _, testT := range tests {
		t.Run(testT.name, func(t *testing.T) {
			result := MergeEnvMaps(testT.input...)
			if !reflect.DeepEqual(result, testT.expected) {
				t.Errorf("MergeEnvMaps() = %v, want %v", result, testT.expected)
			}
		})
	}
}

func TestGetFieldName(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{
			name:     "field with mapstructure tag",
			field:    reflect.StructField{Name: "TestField", PkgPath: "", Type: reflect.TypeOf(""), Tag: `mapstructure:"custom_name"`, Offset: 0, Index: nil, Anonymous: false},
			expected: "custom_name",
		},
		{
			name:     "field without mapstructure tag",
			field:    reflect.StructField{Name: "TestField", PkgPath: "", Type: reflect.TypeOf(""), Tag: "", Offset: 0, Index: nil, Anonymous: false},
			expected: "testfield",
		},
		{
			name:     "field with empty mapstructure tag",
			field:    reflect.StructField{Name: "TestField", PkgPath: "", Type: reflect.TypeOf(""), Tag: `mapstructure:""`, Offset: 0, Index: nil, Anonymous: false},
			expected: "testfield",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := getFieldName(reflect.StructField{
				Name:      testCase.field.Name,
				PkgPath:   testCase.field.PkgPath,
				Type:      testCase.field.Type,
				Tag:       testCase.field.Tag,
				Offset:    testCase.field.Offset,
				Index:     testCase.field.Index,
				Anonymous: testCase.field.Anonymous,
			})
			if result != testCase.expected {
				t.Errorf("getFieldName() = %v, want %v", result, testCase.expected)
			}
		})
	}
}

func TestBuildKey(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		prefix    string
		expected  string
	}{
		{
			name:      "no prefix",
			fieldName: "host",
			prefix:    "",
			expected:  "host",
		},
		{
			name:      "with prefix",
			fieldName: "host",
			prefix:    "database",
			expected:  "database.host",
		},
		{
			name:      "nested prefix",
			fieldName: "name",
			prefix:    "database.connection",
			expected:  "database.connection.name",
		},
	}

	for _, testT := range tests {
		t.Run(testT.name, func(t *testing.T) {
			result := buildKey(testT.fieldName, testT.prefix)
			if result != testT.expected {
				t.Errorf("buildKey() = %v, want %v", result, testT.expected)
			}
		})
	}
}

func TestHandleSliceField(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected []string
		wantErr  bool
	}{
		{
			name:     "string slice with comma-separated values",
			envValue: "value1,value2,value3",
			expected: []string{"value1", "value2", "value3"},
			wantErr:  false,
		},
		{
			name:     "string slice with spaces",
			envValue: "value1, value2 , value3",
			expected: []string{"value1", "value2", "value3"},
			wantErr:  false,
		},
		{
			name:     "single value",
			envValue: "single_value",
			expected: []string{"single_value"},
			wantErr:  false,
		},
		{
			name:     "empty value",
			envValue: "",
			expected: []string{""},
			wantErr:  false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create a new slice to modify
			slicePtr := reflect.New(reflect.SliceOf(reflect.TypeOf("")))
			sliceVal := slicePtr.Elem()

			err := handleSliceField(sliceVal, testCase.envValue)
			if (err != nil) != testCase.wantErr {
				t.Errorf("handleSliceField() error = %v, wantErr %v", err, testCase.wantErr)
				return
			}

			if !testCase.wantErr {
				result, ok := sliceVal.Interface().([]string)
				if !ok {
					t.Errorf("handleSliceField() did not return a []string slice")
					return
				}
				if !reflect.DeepEqual(result, testCase.expected) {
					t.Errorf("handleSliceField() = %v, want %v", result, testCase.expected)
				}
			}
		})
	}
}

func TestHandleSliceFieldNonString(t *testing.T) {
	// Test that non-string slices are not processed
	intSlicePtr := reflect.New(reflect.SliceOf(reflect.TypeOf(0)))
	intSliceVal := intSlicePtr.Elem()

	err := handleSliceField(intSliceVal, "1,2,3")
	if err != nil {
		t.Errorf("handleSliceField() with non-string slice should not error, got %v", err)
	}

	// Should remain empty since it's not a string slice
	if intSliceVal.Len() != 0 {
		t.Error("handleSliceField() should not modify non-string slices")
	}
}

func TestHandleStringField(t *testing.T) {
	var testString string
	val := reflect.ValueOf(&testString).Elem()

	handleStringField(val, "test_value")

	if testString != "test_value" {
		t.Errorf("handleStringField() = %v, want %v", testString, "test_value")
	}
}

func TestFillStructFromEnv(t *testing.T) {
	type NestedStruct struct {
		Host string
		Port string
	}

	type TestStruct struct {
		Name     string
		Database NestedStruct
		Tags     []string
	}

	tests := []struct {
		name     string
		env      map[string]string
		expected TestStruct
	}{
		{
			name: "simple fields",
			env: map[string]string{
				"name": "test_app",
			},
			expected: TestStruct{
				Name:     "test_app",
				Database: NestedStruct{Host: "", Port: ""},
				Tags:     nil,
			},
		},
		{
			name: "nested struct",
			env: map[string]string{
				"name":          "test_app",
				"database.host": "localhost",
				"database.port": "5432",
			},
			expected: TestStruct{
				Name:     "test_app",
				Database: NestedStruct{Host: "localhost", Port: "5432"},
				Tags:     nil,
			},
		},
		{
			name: "slice field",
			env: map[string]string{
				"name": "test_app",
				"tags": "tag1,tag2,tag3",
			},
			expected: TestStruct{
				Name:     "test_app",
				Database: NestedStruct{Host: "", Port: ""},
				Tags:     []string{"tag1", "tag2", "tag3"},
			},
		},
	}

	for _, testT := range tests {
		t.Run(testT.name, func(t *testing.T) {
			var result TestStruct
			err := FillStructFromEnv("", reflect.ValueOf(&result).Elem(), testT.env)
			if err != nil {
				t.Errorf("FillStructFromEnv() error = %v", err)
				return
			}

			if !reflect.DeepEqual(result, testT.expected) {
				t.Errorf("FillStructFromEnv() = %v, want %v", result, testT.expected)
			}
		})
	}
}

func TestFillStructFromEnvWithPointer(t *testing.T) {
	type TestStruct struct {
		Name string
	}

	var result TestStruct
	env := map[string]string{"name": "test_value"}

	// Test with pointer to struct
	err := FillStructFromEnv("", reflect.ValueOf(&result), env)
	if err != nil {
		t.Errorf("FillStructFromEnv() with pointer error = %v", err)
		return
	}

	if result.Name != "test_value" {
		t.Errorf("FillStructFromEnv() with pointer = %v, want %v", result.Name, "test_value")
	}
}

func TestFillStructFromEnvWithNonStruct(t *testing.T) {
	var result string
	env := map[string]string{"test": "value"}

	// Should not error with non-struct types
	err := FillStructFromEnv("", reflect.ValueOf(&result).Elem(), env)
	if err != nil {
		t.Errorf("FillStructFromEnv() with non-struct should not error, got %v", err)
	}
}

func TestFillStructFromEnvWithPrefix(t *testing.T) {
	type TestStruct struct {
		Name string
	}

	var result TestStruct
	env := map[string]string{"prefix.name": "test_value"}

	err := FillStructFromEnv("prefix", reflect.ValueOf(&result).Elem(), env)
	if err != nil {
		t.Errorf("FillStructFromEnv() with prefix error = %v", err)
		return
	}

	if result.Name != "test_value" {
		t.Errorf("FillStructFromEnv() with prefix = %v, want %v", result.Name, "test_value")
	}
}

func TestFillStructFromEnvWithAnonymousStruct(t *testing.T) {
	type EmbeddedStruct struct {
		EmbeddedField string
	}

	type TestStruct struct {
		Name string
		EmbeddedStruct
	}

	var result TestStruct
	env := map[string]string{
		"name":          "test_value",
		"embeddedfield": "embedded_value",
	}

	err := FillStructFromEnv("", reflect.ValueOf(&result).Elem(), env)
	if err != nil {
		t.Errorf("FillStructFromEnv() with anonymous struct error = %v", err)
		return
	}

	if result.Name != "test_value" {
		t.Errorf("FillStructFromEnv() Name = %v, want %v", result.Name, "test_value")
	}

	if result.EmbeddedField != "embedded_value" {
		t.Errorf("FillStructFromEnv() EmbeddedField = %v, want %v", result.EmbeddedField, "embedded_value")
	}
}
