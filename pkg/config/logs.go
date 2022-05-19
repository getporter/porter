package config

import "go.uber.org/zap/zapcore"

// LogConfig are settings related to Porter's log files.
type LogConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Level   LogLevel `mapstructure:"level"`
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
