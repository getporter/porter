package config

import (
	"time"

	"go.uber.org/zap/zapcore"
)

// LogConfig are settings related to Porter's log files.
type LogConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Level   LogLevel `mapstructure:"level"`
	// Structured indicates if the logs sent to the console should include timestamp and log levels
	Structured bool     `mapstructure:"structured"`
}

// TelemetryConfig specifies how to connect to an open telemetry collector.
// See https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace
type TelemetryConfig struct {
	Enabled     bool              `mapstructure:"enabled"`
	Endpoint    string            `mapstructure:"endpoint"`
	Protocol    string            `mapstructure:"protocol"`
	Insecure    bool              `mapstructure:"insecure"`
	Certificate string            `mapstructure:"certificate"`
	Headers     map[string]string `mapstructure:"headers"`
	Timeout     string            `mapstructure:"timeout"`
	Compression string            `mapstructure:"compression"`

	// RedirectToFile instructs Porter to write telemetry data to a file in
	// PORTER_HOME/traces instead of exporting it to a collector
	RedirectToFile bool `mapstructure:"redirect-to-file"`

	// StartTimeout sets the amount of time to wait while establishing a connection
	// to the OpenTelemetry collector.
	StartTimeout string `mapstructure:"start-timeout"`
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

func (l LogLevel) Level() zapcore.Level {
	switch l {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
