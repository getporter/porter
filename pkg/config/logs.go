package config

import "go.uber.org/zap/zapcore"

// LogConfig are settings related to Porter's log files.
type LogConfig struct {
	Enabled bool     `mapstructure:"enabled,omitempty"`
	Level   LogLevel `mapstructure:"level,omitempty"`
}

// TelemetryConfig specifies how to connect to an open telemetry collector.
// See https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace
type TelemetryConfig struct {
	Enabled     bool              `mapstructure:"enabled,omitempty"`
	Endpoint    string            `mapstructure:"endpoint,omitempty"`
	Protocol    string            `mapstructure:"protocol,omitempty"`
	Insecure    bool              `mapstructure:"insecure,omitempty"`
	Certificate string            `mapstructure:"certificate,omitempty"`
	Headers     map[string]string `mapstructure:"headers,omitempty"`
	Timeout     string            `mapstructure:"timeout,omitempty"`
	Compression string            `mapstructure:"compression,omitempty"`
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
