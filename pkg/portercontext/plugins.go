package portercontext

import (
	"encoding/json"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// makePluginLogger creates a logger suitable for plugins to communicate with the hashicorp
// go-plugin framework, using hclog to talk over stderr
func (c *Context) makePluginLogger(pluginKey string, cfg LogConfiguration) zapcore.Core {
	pluginLogger := zapToHclog{
		hclog.New(&hclog.LoggerOptions{
			Name:       pluginKey,
			Output:     c.Err,
			Level:      hclog.Debug,
			JSONFormat: true,
		}),
	}

	enc := zap.NewProductionEncoderConfig()
	jsonEncoder := zapcore.NewJSONEncoder(enc)
	return zapcore.NewCore(jsonEncoder, zapcore.AddSync(pluginLogger), cfg.LogLevel)
}

// Accepts zap log commands and translates them to a format that hclog understands
// so that the plugin doesn't write any log messages that would cause the go plugin
// framework to error out (i.e. printing directly to stderr)
type zapToHclog struct {
	logger hclog.Logger
}

func (z zapToHclog) Write(p []byte) (n int, err error) {
	var entry map[string]interface{}
	if err := json.Unmarshal(p, &entry); err != nil {
		return 0, err
	}
	msg := entry["msg"].(string)

	switch entry["level"].(string) {
	case "error":
		z.logger.Error(msg)
	case "warn":
		z.logger.Warn(msg)
	case "debug":
		z.logger.Debug(msg)
	default:
		z.logger.Info(msg)
	}
	return len(p), nil
}
