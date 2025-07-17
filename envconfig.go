// Package envconfig provides generic environment variable handling and configuration
// loading utilities that can be used across different Go projects.
//
// The package supports:
// - Parsing .env files from byte arrays
// - Loading system environment variables
// - Merging multiple environment sources with priority
// - Filling struct fields from environment variables using reflection
// - Validation of configuration structs using struct tags
package envconfig

import (
	"bufio"
	"bytes"
	"os"
	"reflect"
	"strings"
)

// ReadDotenvBytes parses environment variables from a .env file content provided as bytes.
// It returns a map of key-value pairs where keys are environment variable names and values
// are their corresponding values.
//
// The function handles:
// - Empty lines (ignored)
// - Key-value pairs separated by = (first = is used as separator)
// - Whitespace trimming for both keys and values
//
// Example:
//
//	data := []byte("DB_HOST=localhost\nDB_PORT=5432\n# comment\nEMPTY_VAR=")
//	envVars := ReadDotenvBytes(data)
//	// Returns: map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432", "EMPTY_VAR": ""}
func ReadDotenvBytes(data []byte) map[string]string {
	out := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) == 2 {
			out[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return out
}

// GetEnvs retrieves all system environment variables and returns them as a map.
// It iterates through all environment variables available to the current process
// and returns them as key-value pairs.
//
// The function handles:
// - All system environment variables
// - Variables with empty values (included in result)
// - Proper parsing of KEY=VALUE format
//
// Example:
//
//	envVars := GetEnvs()
//	// Returns: map[string]string{"PATH": "/usr/bin:/bin", "HOME": "/home/user", ...}
func GetEnvs() map[string]string {
	out := map[string]string{}
	for _, env := range os.Environ() {
		kv := strings.SplitN(env, "=", 2)
		if len(kv) == 2 {
			out[kv[0]] = kv[1]
		}
	}
	return out
}

// MergeEnvMaps combines multiple environment variable maps into a single map.
// Later maps override earlier ones, and keys are normalized to lowercase with
// underscores converted to dots for consistent access patterns.
//
// The function:
// - Processes maps in order (later maps have priority)
// - Normalizes keys to lowercase
// - Converts underscores to dots (DATABASE_HOST -> database.host)
// - Skips empty values
// - Returns a single merged map
//
// Example:
//
//	map1 := map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432"}
//	map2 := map[string]string{"DB_HOST": "prod-server", "API_KEY": "secret"}
//	merged := MergeEnvMaps(map1, map2)
//	// Returns: map[string]string{"db.host": "prod-server", "db.port": "5432", "api.key": "secret"}
func MergeEnvMaps(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			norm := strings.ToLower(strings.ReplaceAll(k, "_", "."))
			if v != "" {
				out[norm] = v
			}
		}
	}
	return out
}

// getFieldName extracts the field name for environment variable mapping.
// It first checks for a "mapstructure" tag, and if not found, uses the
// lowercase version of the field name.
func getFieldName(field reflect.StructField) string {
	fieldName := field.Tag.Get("mapstructure")
	if fieldName == "" {
		fieldName = strings.ToLower(field.Name)
	}
	return fieldName
}

// buildKey constructs a nested key for environment variable lookups.
// If a prefix is provided, it combines the prefix and field name with a dot.
// Otherwise, it returns just the field name.
func buildKey(fieldName, prefix string) string {
	if prefix != "" {
		return prefix + "." + fieldName
	}
	return fieldName
}

// handleSliceField processes slice fields from environment variables.
// It only handles string slices, splitting the environment value by commas
// and creating a new slice with trimmed values.
func handleSliceField(val reflect.Value, envValue string) error {
	if val.Type().Elem().Kind() != reflect.String {
		return nil
	}
	parts := strings.Split(envValue, ",")
	slice := reflect.MakeSlice(val.Type(), len(parts), len(parts))
	for i, part := range parts {
		slice.Index(i).SetString(strings.TrimSpace(part))
	}
	val.Set(slice)
	return nil
}

// handleStringField sets a string field value from an environment variable.
func handleStringField(val reflect.Value, envValue string) {
	val.SetString(envValue)
}

// handleStructField processes nested struct fields by recursively calling FillStructFromEnv.
func handleStructField(val reflect.Value, key string, env map[string]string) error {
	return FillStructFromEnv(key, val, env)
}

// processStructField processes a single struct field, determining its type and
// calling the appropriate handler function. It supports structs, strings, and slices.
func processStructField(field reflect.StructField, val reflect.Value, prefix string, env map[string]string) error {
	if !val.CanSet() {
		return nil
	}

	fieldName := getFieldName(field)
	key := buildKey(fieldName, prefix)

	switch val.Kind() {
	case reflect.Struct:
		return handleStructField(val, key, env)
	case reflect.String:
		if s, ok := env[key]; ok {
			handleStringField(val, s)
		}
		return nil
	case reflect.Slice:
		if s, ok := env[key]; ok {
			return handleSliceField(val, s)
		}
		return nil
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr,
		reflect.UnsafePointer:
		return nil
	default:
		return nil
	}
}

// FillStructFromEnv fills a struct's fields with values from environment variables.
// It uses reflection to iterate through struct fields and populate them based on
// environment variable mappings.
//
// The function supports:
// - Nested structs (with dot notation in keys)
// - String fields
// - String slices (comma-separated values)
// - Anonymous embedded structs
// - Pointer dereferencing
//
// Environment variable keys are matched against struct field names using the
// following rules:
// - Field names are converted to lowercase
// - Nested structs use dot notation (e.g., "database.host")
// - mapstructure tags override default field names
//
// Example:
//
//	type Config struct {
//	  Host string
//	  Database struct {
//	    Host string
//	    Port int
//	  }
//	}
//
//	env := map[string]string{
//	  "host": "localhost",
//	  "database.host": "db-server",
//	}
//
//	config := &Config{}
//	err := FillStructFromEnv("", reflect.ValueOf(config).Elem(), env)
func FillStructFromEnv(prefix string, value reflect.Value, env map[string]string) error {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil
	}

	t := value.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		val := value.Field(i)
		if field.Anonymous && val.Kind() == reflect.Struct {
			if err := FillStructFromEnv(prefix, val, env); err != nil {
				return err
			}
			continue
		}
		if err := processStructField(field, val, prefix, env); err != nil {
			return err
		}
	}
	return nil
}
