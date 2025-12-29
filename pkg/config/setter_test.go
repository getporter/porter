package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetConfigValue_StringField(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "verbosity", "debug")
	require.NoError(t, err)
	assert.Equal(t, "debug", data.Verbosity)
}

func TestSetConfigValue_BoolField(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &Data{}
			err := SetConfigValue(data, "allow-docker-host-access", tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, data.AllowDockerHostAccess)
		})
	}
}

func TestSetConfigValue_NestedField_OneLevel(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "logs.level", "warn")
	require.NoError(t, err)
	assert.Equal(t, LogLevel("warn"), data.Logs.Level)
}

func TestSetConfigValue_NestedField_BoolType(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "telemetry.enabled", "true")
	require.NoError(t, err)
	assert.True(t, data.Telemetry.Enabled)
}

func TestSetConfigValue_InvalidPath_UnknownField(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "nonexistent", "value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config field")
}

func TestSetConfigValue_InvalidPath_UnknownNestedField(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "logs.nonexistent", "value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config field")
}

func TestSetConfigValue_InvalidPath_NavigateThroughNonStruct(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "verbosity.nested.field", "value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not a struct")
}

func TestSetConfigValue_InvalidValue_EmptyPath(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "", "value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config path cannot be empty")
}

func TestSetConfigValue_InvalidValue_EmptyValue(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "verbosity", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config value cannot be empty")
}

func TestSetConfigValue_InvalidValue_BoolParsing(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "allow-docker-host-access", "notabool")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid boolean value")
}

func TestSetConfigValue_PreservesOtherFields(t *testing.T) {
	data := &Data{
		Verbosity: "info",
		Namespace: "original",
	}

	err := SetConfigValue(data, "verbosity", "debug")
	require.NoError(t, err)

	assert.Equal(t, "debug", data.Verbosity)
	assert.Equal(t, "original", data.Namespace, "other fields should be preserved")
}

func TestSetConfigValue_MultipleUpdates(t *testing.T) {
	data := &Data{}

	require.NoError(t, SetConfigValue(data, "verbosity", "debug"))
	require.NoError(t, SetConfigValue(data, "namespace", "test"))
	require.NoError(t, SetConfigValue(data, "logs.structured", "true"))

	assert.Equal(t, "debug", data.Verbosity)
	assert.Equal(t, "test", data.Namespace)
	assert.True(t, data.Logs.Structured)
}

func TestValidateConfigValue_ValidPath(t *testing.T) {
	data := &Data{}
	err := ValidateConfigValue(data, "verbosity", "debug")
	assert.NoError(t, err)
}

func TestValidateConfigValue_ValidNestedPath(t *testing.T) {
	data := &Data{}
	err := ValidateConfigValue(data, "logs.level", "info")
	assert.NoError(t, err)
}

func TestValidateConfigValue_InvalidPath(t *testing.T) {
	data := &Data{}
	err := ValidateConfigValue(data, "nonexistent", "value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config field")
}

func TestValidateConfigValue_InvalidBoolValue(t *testing.T) {
	data := &Data{}
	err := ValidateConfigValue(data, "allow-docker-host-access", "notabool")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid boolean value")
}

func TestValidateConfigValue_EmptyPath(t *testing.T) {
	data := &Data{}
	err := ValidateConfigValue(data, "", "value")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config path cannot be empty")
}

func TestValidateConfigValue_EmptyValue(t *testing.T) {
	data := &Data{}
	err := ValidateConfigValue(data, "verbosity", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config value cannot be empty")
}

func TestSetConfigValue_DomainValidation_LogLevel(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		shouldError bool
	}{
		{"valid debug", "debug", false},
		{"valid info", "info", false},
		{"valid warn", "warn", false},
		{"valid warning", "warning", false},
		{"valid error", "error", false},
		{"invalid value", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &Data{}
			err := SetConfigValue(data, "logs.level", tt.value)
			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid value")
			} else {
				require.NoError(t, err)
				assert.Equal(t, LogLevel(tt.value), data.Logs.Level)
			}
		})
	}
}

func TestSetConfigValue_DomainValidation_RuntimeDriver(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		shouldError bool
	}{
		{"valid docker", "docker", false},
		{"valid kubernetes", "kubernetes", false},
		{"invalid value", "podman", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &Data{}
			err := SetConfigValue(data, "runtime-driver", tt.value)
			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid value")
				assert.Contains(t, err.Error(), "docker, kubernetes")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.value, data.RuntimeDriver)
			}
		})
	}
}

func TestSetConfigValue_DomainValidation_BuildDriver(t *testing.T) {
	data := &Data{}
	err := SetConfigValue(data, "build-driver", "buildkit")
	require.NoError(t, err)
	assert.Equal(t, "buildkit", data.BuildDriver)

	err = SetConfigValue(data, "build-driver", "docker")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value")
}

func TestSetConfigValue_DomainValidation_TelemetryProtocol(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		shouldError bool
	}{
		{"valid grpc", "grpc", false},
		{"valid http/protobuf", "http/protobuf", false},
		{"invalid http", "http", true},
		{"invalid value", "tcp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &Data{}
			err := SetConfigValue(data, "telemetry.protocol", tt.value)
			if tt.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid value")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.value, data.Telemetry.Protocol)
			}
		})
	}
}

func TestSetConfigValue_DomainValidation_Verbosity(t *testing.T) {
	tests := []struct {
		value       string
		shouldError bool
	}{
		{"debug", false},
		{"info", false},
		{"warn", false},
		{"warning", false},
		{"error", false},
		{"trace", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			data := &Data{}
			err := SetConfigValue(data, "verbosity", tt.value)
			if tt.shouldError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFindFieldByTagWithMeta_NonStructValue(t *testing.T) {
	// Test that findFieldByTagWithMeta returns an error when given a non-struct value
	// This prevents a panic from calling NumField() on non-struct types
	tests := []struct {
		name  string
		value interface{}
	}{
		{"string", "test"},
		{"int", 42},
		{"bool", true},
		{"slice", []string{"a", "b"}},
		{"map", map[string]string{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			typ := v.Type()
			_, _, err := findFieldByTagWithMeta(v, typ, "anyfield")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected struct")
		})
	}
}
