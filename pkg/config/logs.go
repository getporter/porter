package config

import (
	"time"

	"go.uber.org/zap/zapcore"
)

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogConfig are settings related to Porter's log files.
type LogConfig struct {
	// Structured indicates if the logs sent to the console should include timestamp and log levels
	Structured bool     `mapstructure:"structured" toml:"structured" yaml:"structured" json:"structured"`
	LogToFile  bool     `mapstructure:"log-to-file" toml:"log-to-file" yaml:"log-to-file" json:"log-to-file"`
	Level      LogLevel `mapstructure:"level" toml:"level" yaml:"level" json:"level"`
}

// TelemetryConfig specifies how to connect to an open telemetry collector.
// See https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace
type TelemetryConfig struct {
	Enabled     bool              `mapstructure:"enabled" toml:"enabled" yaml:"enabled" json:"enabled"`
	Endpoint    string            `mapstructure:"endpoint" toml:"endpoint" yaml:"endpoint" json:"endpoint"`
	Protocol    string            `mapstructure:"protocol" toml:"protocol" yaml:"protocol" json:"protocol"`
	Insecure    bool              `mapstructure:"insecure" toml:"insecure" yaml:"insecure" json:"insecure"`
	Certificate string            `mapstructure:"certificate" toml:"certificate" yaml:"certificate" json:"certificate"`
	Headers     map[string]string `mapstructure:"headers" toml:"headers" yaml:"headers" json:"headers"`
	Timeout     string            `mapstructure:"timeout" toml:"timeout" yaml:"timeout" json:"timeout"`
	Compression string            `mapstructure:"compression" toml:"compression" yaml:"compression" json:"compression"`

	// RedirectToFile instructs Porter to write telemetry data to a file in
	// PORTER_HOME/traces instead of exporting it to a collector
	RedirectToFile bool `mapstructure:"redirect-to-file" toml:"redirect-to-file" yaml:"redirect-to-file" json:"redirect-to-file"`

	// StartTimeout sets the amount of time to wait while establishing a connection
	// to the OpenTelemetry collector.
	StartTimeout string `mapstructure:"start-timeout" toml:"start-timeout" yaml:"start-timeout" json:"start-timeout"`
}

// GetStartTimeout returns the amount of time to wait for the collector to start
// if a value was not configured, return the default timeout.
func (c TelemetryConfig) GetStartTimeout() time.Duration {
	if timeout, err := time.ParseDuration(c.StartTimeout); err == nil {
		return timeout
	}
	return 100 * time.Millisecond
}

type LogLevel string

// ParseLogLevel reads the string representation of a LogLevel and converts it to a LogLevel.
// Unrecognized values default to info.
func ParseLogLevel(value string) LogLevel {
	switch value {
	case "debug", "info", "warn", "error":
		return LogLevel(value)
	case "warning": // be nice to people who can't type
		return "warn"
	default:
		return "info"
	}
}

func (l LogLevel) Level() zapcore.Level {
	switch l {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
