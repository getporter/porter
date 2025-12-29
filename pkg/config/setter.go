package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// SetConfigValue sets a field on the Data struct using dot notation path.
// Supports simple fields and nested fields like "logs.level" or "telemetry.enabled".
// Uses reflection to dynamically set fields based on struct tags.
func SetConfigValue(data *Data, path string, value string) error {
	if err := ValidateConfigValue(data, path, value); err != nil {
		return err
	}

	parts := strings.Split(path, ".")
	return setFieldByPath(reflect.ValueOf(data).Elem(), parts, value)
}

// setFieldByPath recursively navigates and sets a field using the path segments
func setFieldByPath(v reflect.Value, parts []string, value string) error {
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}

	fieldName := parts[0]
	field, err := findFieldByTag(v, fieldName)
	if err != nil {
		return err
	}

	// If this is the last part, set the value
	if len(parts) == 1 {
		return setFieldValue(field, value)
	}

	// Navigate to nested struct
	if field.Kind() != reflect.Struct {
		return fmt.Errorf("field %s is not a struct, cannot navigate deeper", fieldName)
	}

	return setFieldByPath(field, parts[1:], value)
}

// findFieldByTag finds a struct field by matching against toml/yaml/json tags
func findFieldByTag(v reflect.Value, name string) (reflect.Value, error) {
	field, _, err := findFieldByTagWithMeta(v, v.Type(), name)
	return field, err
}

// findFieldByTagWithMeta finds a struct field and returns both the value and StructField metadata
func findFieldByTagWithMeta(v reflect.Value, t reflect.Type, name string) (reflect.Value, reflect.StructField, error) {
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, reflect.StructField{}, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Check toml, yaml, and json tags
		for _, tagKey := range []string{"toml", "yaml", "json"} {
			if tagValue := field.Tag.Get(tagKey); tagValue != "" {
				// Handle tags like "field-name,omitempty"
				tagName := strings.Split(tagValue, ",")[0]
				if tagName == name {
					fieldValue := v.Field(i)
					if !fieldValue.CanSet() {
						return reflect.Value{}, reflect.StructField{}, fmt.Errorf("field %s cannot be set", name)
					}
					return fieldValue, field, nil
				}
			}
		}
	}

	return reflect.Value{}, reflect.StructField{}, fmt.Errorf("unknown config field: %s", name)
}

// setFieldValue sets a reflect.Value to the given string value with type conversion
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return nil

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s (valid values: true, false, 1, 0)", value)
		}
		field.SetBool(boolVal)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
		field.SetInt(intVal)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value: %s", value)
		}
		field.SetUint(uintVal)
		return nil

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(floatVal)
		return nil

	default:
		return fmt.Errorf("unsupported field type: %s (only string, bool, and numeric types are supported)", field.Kind())
	}
}

// ValidateConfigValue validates that a config path exists and the value is valid for its type
func ValidateConfigValue(data *Data, path string, value string) error {
	if path == "" {
		return fmt.Errorf("config path cannot be empty")
	}

	if value == "" {
		return fmt.Errorf("config value cannot be empty")
	}

	parts := strings.Split(path, ".")

	// Validate the path exists by attempting to find the field
	v := reflect.ValueOf(data).Elem()
	t := v.Type()

	for i, part := range parts {
		field, sf, err := findFieldByTagWithMeta(v, t, part)
		if err != nil {
			// Provide helpful context about where we are in the path
			if i > 0 {
				return fmt.Errorf("invalid config path %s: %w (at level %d)", path, err, i+1)
			}
			return err
		}

		// If this is the last part, validate the value can be set
		if i == len(parts)-1 {
			return validateValueForTypeWithTag(field, sf, value, path)
		}

		// Navigate to nested struct
		if field.Kind() != reflect.Struct {
			return fmt.Errorf("field %s is not a struct, cannot navigate to %s", part, parts[i+1])
		}

		v = field
		t = field.Type()
	}

	return nil
}

// validateValueForTypeWithTag checks if a value can be converted to the field's type
// and validates against any constraints defined in the validate tag
func validateValueForTypeWithTag(field reflect.Value, sf reflect.StructField, value string, path string) error {
	// First check type conversion
	switch field.Kind() {
	case reflect.String:
		// Check validate tag for string constraints
		return validateWithTag(sf, value, path)

	case reflect.Bool:
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("invalid boolean value for %s: %s (valid values: true, false, 1, 0)", path, value)
		}
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("invalid integer value for %s: %s", path, value)
		}
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if _, err := strconv.ParseUint(value, 10, 64); err != nil {
			return fmt.Errorf("invalid unsigned integer value for %s: %s", path, value)
		}
		return nil

	case reflect.Float32, reflect.Float64:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("invalid float value for %s: %s", path, value)
		}
		return nil

	default:
		return fmt.Errorf("field %s has unsupported type %s (only string, bool, and numeric types can be set)", path, field.Kind())
	}
}

// validateWithTag validates a value against the validate struct tag
func validateWithTag(sf reflect.StructField, value string, path string) error {
	validateTag := sf.Tag.Get("validate")
	if validateTag == "" {
		return nil // No validation tag
	}

	// Parse the validate tag - currently only support "oneof=value1 value2 ..."
	if strings.HasPrefix(validateTag, "oneof=") {
		validValues := strings.Split(strings.TrimPrefix(validateTag, "oneof="), " ")
		for _, valid := range validValues {
			if value == valid {
				return nil
			}
		}
		return fmt.Errorf("invalid value for %s: %s (valid values: %s)", path, value, strings.Join(validValues, ", "))
	}

	return nil
}
