package portercontext

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/carolynvs/aferox"
	cnabclaims "github.com/cnabio/cnab-go/claim"
	"github.com/hashicorp/go-multierror"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// MixinOutputsDir represents the directory where mixin output files are written/read
	MixinOutputsDir = "/cnab/app/porter/outputs"

	// EnvCorrelationID is the name of the environment variable containing the
	// id to correlate logs with a workflow.
	EnvCorrelationID = "PORTER_CORRELATION_ID"
)

type CommandBuilder func(ctx context.Context, name string, arg ...string) *exec.Cmd

type Context struct {
	environ            map[string]string
	FileSystem         aferox.Aferox
	In                 io.Reader
	Out                io.Writer
	Err                io.Writer
	NewCommand         CommandBuilder
	PlugInDebugContext *PluginDebugContext

	//
	// Logging and Tracing configuration
	//

	// a consistent id that is set on the context and emitted in the logs
	// Helps correlate logs with a workflow.
	correlationId string

	// logCfg is the logger configuration used.
	logCfg LogConfiguration

	// logFile is the open file where we are sending logs.
	logFile afero.File

	// traceFile is the open file where we are sending traces.
	traceFile afero.File

	// indicates if log timestamps should be printed to the console
	timestampLogs bool

	// handles sending tracing data to an otel collector
	tracer tracing.Tracer

	// indicates if we have created a real tracer yet (instead of noop)
	tracerInitalized bool

	// handles send log data to the console/logfile
	logger *zap.Logger

	// IsInternalPlugin indicates that Porter is running as an internal plugin
	IsInternalPlugin bool

	// InternalPluginKey is the current plugin that Porter is running as, e.g. storage.porter.mongodb
	InternalPluginKey string
}

// New creates a new context in the specified directory.
func New() *Context {
	// Ignore any error getting the working directory and report errors
	// when we attempt to access files in the current directory. This
	// allows us to use the current directory as a default, and allow
	// tests to override it.
	pwd, _ := os.Getwd()

	correlationId := os.Getenv(EnvCorrelationID)
	if correlationId == "" {
		correlationId = cnabclaims.MustNewULID() // not using cnab package because that creates a cycle
	}

	c := &Context{
		environ:       getEnviron(),
		FileSystem:    aferox.NewAferox(pwd, afero.NewOsFs()),
		In:            os.Stdin,
		Out:           NewCensoredWriter(os.Stdout),
		Err:           NewCensoredWriter(os.Stderr),
		correlationId: correlationId,
		timestampLogs: true,
	}

	// Make the correlation id available for the plugins to use
	c.Setenv(EnvCorrelationID, correlationId)

	c.ConfigureLogging(context.Background(), LogConfiguration{})
	c.defaultNewCommand()
	c.PlugInDebugContext = NewPluginDebugContext(c)

	return c
}

// StartRootSpan creates the root tracing span for the porter application.
// This should only be done once.
func (c *Context) StartRootSpan(ctx context.Context, op string, attrs ...attribute.KeyValue) (context.Context, tracing.RootTraceLogger) {
	childCtx, span := c.tracer.Start(ctx, op)
	attrs = append(attrs, attribute.String("correlation-id", c.correlationId))
	span.SetAttributes(attrs...)
	return tracing.NewRootLogger(childCtx, span, c.logger, c.tracer)
}

func (c *Context) makeLogEncoding() zapcore.EncoderConfig {
	enc := zap.NewProductionEncoderConfig()
	if c.timestampLogs {
		enc.EncodeTime = zapcore.ISO8601TimeEncoder
	} else { // used for testing, so we don't have unique timestamps in the logs
		enc.EncodeTime = func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString("")
		}
	}
	return enc
}

type LogConfiguration struct {
	// Verbosity is the threshold for printing messages to the console.
	Verbosity zapcore.Level

	LogToFile    bool
	LogDirectory string

	// LogLevel is the threshold for writing messages to the log file.
	LogLevel                zapcore.Level
	StructuredLogs          bool
	TelemetryEnabled        bool
	TelemetryEndpoint       string
	TelemetryProtocol       string
	TelemetryInsecure       bool
	TelemetryCertificate    string
	TelemetryCompression    string
	TelemetryTimeout        string
	TelemetryHeaders        map[string]string
	TelemetryServiceName    string
	TelemetryDirectory      string
	TelemetryRedirectToFile bool
	TelemetryStartTimeout   time.Duration
}

// ConfigureLogging applies different configuration to our logging and tracing.
func (c *Context) ConfigureLogging(ctx context.Context, cfg LogConfiguration) {
	c.logCfg = cfg

	var baseLogger zapcore.Core
	if c.IsInternalPlugin {
		c.logCfg.TelemetryServiceName = c.InternalPluginKey
		baseLogger = c.makePluginLogger(c.InternalPluginKey, cfg)
	} else {
		baseLogger = c.makeConsoleLogger()
	}

	c.configureLoggingWith(ctx, baseLogger)
}

// ConfigureLogging applies different configuration to our logging and tracing.
func (c *Context) configureLoggingWith(ctx context.Context, baseLogger zapcore.Core) {
	// make a temporary logger that we can use until we've completely initialized the full logger
	tmpLog := zap.New(baseLogger)

	var err error
	fileLogger := zapcore.NewNopCore()
	if c.logCfg.LogToFile {
		fileLogger, err = c.configureFileLog(c.logCfg.LogDirectory)
		if err != nil {
			tmpLog.Error(fmt.Errorf("could not configure a file logger: %w", err).Error())
		} else {
			tmpLog.Debug("Writing logs to " + c.logFile.Name())
		}
	}
	tmpLog = zap.New(zapcore.NewTee(baseLogger, fileLogger))

	if c.logCfg.TelemetryEnabled {
		// Only initialize the tracer once per command
		if !c.tracerInitalized {
			err = c.configureTelemetry(ctx, c.logCfg, tmpLog)
			if err != nil {
				tmpLog.Error(fmt.Errorf("could not configure a tracer: %w", err).Error())
			}
		}
	} else {
		tracer := createNoopTracer()
		c.tracer = tracer
	}

	c.logger = tmpLog
}

func (c *Context) makeConsoleLogger() zapcore.Core {
	encoding := c.makeLogEncoding()

	stderr := c.Err
	if f, ok := stderr.(*os.File); ok {
		if isatty.IsTerminal(f.Fd()) {
			stderr = colorable.NewColorable(f)
			encoding.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		}
	}

	// if structured-logs feature isn't enabled, keep the logs looking like they do now, with just the message printed
	if !c.logCfg.StructuredLogs {
		encoding.TimeKey = ""
		encoding.LevelKey = ""
	}
	consoleEncoder := zapcore.NewConsoleEncoder(encoding)
	return zapcore.NewCore(consoleEncoder, zapcore.AddSync(stderr), c.logCfg.Verbosity)
}

func (c *Context) configureFileLog(dir string) (zapcore.Core, error) {
	if err := c.FileSystem.MkdirAll(dir, pkg.FileModeDirectory); err != nil {
		return nil, err
	}

	// Write the logs to a file
	logfile := filepath.Join(dir, c.buildLogFileName())
	if c.logFile == nil { // We may have already opened this logfile, and we are just changing the log level
		f, err := c.FileSystem.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, pkg.FileModeWritable)
		if err != nil {
			return zapcore.NewNopCore(), fmt.Errorf("could not start log file at %s: %w", logfile, err)
		}
		c.logFile = f
	}

	// Split logs to the console and file
	fileEncoder := zapcore.NewJSONEncoder(c.makeLogEncoding())
	return zapcore.NewCore(fileEncoder, zapcore.AddSync(c.logFile), c.logCfg.LogLevel), nil
}

func (c *Context) buildLogFileName() string {
	// Send plugin logs and traces to a separate file so that there aren't conflicts while writing
	if c.IsInternalPlugin {
		return fmt.Sprintf("%s-%s.json", c.correlationId, c.InternalPluginKey)
	}
	return fmt.Sprintf("%s.json", c.correlationId)
}

func (c *Context) Close() error {
	var bigErr *multierror.Error

	if err := c.tracer.Close(context.Background()); err != nil {
		err = fmt.Errorf("error closing tracer: %w", err)
		bigErr = multierror.Append(bigErr, err)
	}

	if c.traceFile != nil {
		if err := c.traceFile.Close(); err != nil {
			err = fmt.Errorf("error closing trace file %s: %w", c.traceFile.Name(), err)
			bigErr = multierror.Append(bigErr, err)
		}
		c.traceFile = nil
	}

	if c.logFile != nil {
		if err := c.logFile.Close(); err != nil {
			err = fmt.Errorf("error closing log file %s: %w", c.logFile.Name(), err)
			bigErr = multierror.Append(bigErr, err)
		}
		c.logFile = nil
	}

	return bigErr.ErrorOrNil()
}

func (c *Context) defaultNewCommand() {
	c.NewCommand = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return c.CommandContext(ctx, name, arg...)
	}
}

// CommandContext creates a new exec.Cmd in the current directory.
// The provided context is used to kill the process (by calling
// os.Process.Kill) if the context becomes done before the command
// completes on its own.
func (c *Context) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	if filepath.Base(name) == name {
		if lp, ok := c.LookPath(name); ok {
			name = lp
		}
	}

	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.Dir = c.Getwd()
	cmd.Env = c.Environ()
	return cmd
}

func getEnviron() map[string]string {
	environ := map[string]string{}
	for _, env := range os.Environ() {
		envParts := strings.SplitN(env, "=", 2)
		key := envParts[0]
		value := ""
		if len(envParts) > 1 {
			value = envParts[1]
		}
		environ[key] = value
	}
	return environ
}

// Environ returns a copy of strings representing the environment,
// in the form "key=value".
func (c *Context) Environ() []string {
	e := make([]string, 0, len(c.environ))
	for k, v := range c.environ {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	return e
}

// EnvironMap returns a map of the current environment variables.
func (c *Context) EnvironMap() map[string]string {
	env := make(map[string]string, len(c.environ))
	for k, v := range c.environ {
		env[k] = v
	}
	return env
}

// ExpandEnv replaces ${var} or $var in the string according to the values
// of the current environment variables. References to undefined
// variables are replaced by the empty string.
func (c *Context) ExpandEnv(s string) string {
	return os.Expand(s, func(key string) string { return c.Getenv(key) })
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not present.
// To distinguish between an empty value and an unset value, use LookupEnv.
func (c *Context) Getenv(key string) string {
	return c.environ[key]
}

// This is a simplified exec.LookPath that checks if command is accessible given
// a PATH environment variable.
func (c *Context) LookPath(file string) (string, bool) {
	return c.FileSystem.LookPath(file, c.Getenv("PATH"), c.Getenv("PATHEXT"))
}

// LookupEnv retrieves the value of the environment variable named
// by the key. If the variable is present in the environment the
// value (which may be empty) is returned and the boolean is true.
// Otherwise the returned value will be empty and the boolean will
// be false.
func (c *Context) LookupEnv(key string) (string, bool) {
	value, ok := c.environ[key]
	return value, ok
}

// Setenv sets the value of the environment variable named by the key.
// It returns an error, if any.
func (c *Context) Setenv(key string, value string) {
	if c.environ == nil {
		c.environ = make(map[string]string, 1)
	}

	c.environ[key] = value
}

// Unsetenv unsets a single environment variable.
func (c *Context) Unsetenv(key string) {
	delete(c.environ, key)
}

// Clearenv deletes all environment variables.
func (c *Context) Clearenv() {
	c.environ = make(map[string]string, 0)
}

// Getwd returns a rooted path name corresponding to the current directory.
func (c *Context) Getwd() string {
	return c.FileSystem.Getwd()
}

// Chdir changes the current working directory to the named directory.
func (c *Context) Chdir(dir string) {
	c.FileSystem.Chdir(dir)
}

// CensoredWriter is a writer wrapping the provided io.Writer with logic to censor certain values
type CensoredWriter struct {
	writer          io.Writer
	sensitiveValues []string
}

// NewCensoredWriter returns a new CensoredWriter
func NewCensoredWriter(writer io.Writer) *CensoredWriter {
	return &CensoredWriter{writer: writer, sensitiveValues: []string{}}
}

// SetSensitiveValues sets values needing masking for an CensoredWriter
func (cw *CensoredWriter) SetSensitiveValues(vals []string) {
	cw.sensitiveValues = vals
}

// Write implements io.Writer's Write method, performing necessary auditing while doing so
func (cw *CensoredWriter) Write(b []byte) (int, error) {
	auditedBytes := b
	for _, val := range cw.sensitiveValues {
		if strings.TrimSpace(val) != "" {
			auditedBytes = bytes.Replace(auditedBytes, []byte(val), []byte("*******"), -1)
		}
	}

	_, err := cw.writer.Write(auditedBytes)
	return len(b), err
}

func (c *Context) CopyDirectory(srcDir, destDir string, includeBaseDir bool) error {
	var stripPrefix string
	if includeBaseDir {
		stripPrefix = filepath.Dir(srcDir)
	} else {
		stripPrefix = srcDir
	}

	return c.FileSystem.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Translate the path from the src to the final destination
		dest := filepath.Join(destDir, strings.TrimPrefix(path, stripPrefix))
		if dest == "" {
			return nil
		}

		if info.IsDir() {
			err := c.FileSystem.MkdirAll(dest, info.Mode())
			if err != nil {
				return err
			}

			return nil
		}

		return c.CopyFile(path, dest)
	})
}

func (c *Context) CopyFile(src, dest string) error {
	info, err := c.FileSystem.Stat(src)
	if err != nil {
		return err
	}

	data, err := c.FileSystem.ReadFile(src)
	if err != nil {
		return err
	}

	err = c.FileSystem.WriteFile(dest, data, info.Mode())
	if err != nil {
		return err
	}

	return nil
}

// WriteMixinOutputToFile writes the provided bytes (representing a mixin output)
// to a file named by the provided filename in Porter's mixin outputs directory
func (c *Context) WriteMixinOutputToFile(filename string, bytes []byte) error {
	exists, err := c.FileSystem.DirExists(MixinOutputsDir)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.FileSystem.MkdirAll(MixinOutputsDir, pkg.FileModeDirectory); err != nil {
			return fmt.Errorf("couldn't make output directory: %w", err)
		}
	}

	return c.FileSystem.WriteFile(filepath.Join(MixinOutputsDir, filename), bytes, pkg.FileModeWritable)
}

// SetSensitiveValues sets the sensitive values needing masking on output/err streams
// WARNING: This does not work if you are writing to the TraceLogger.
// See https://github.com/getporter/porter/issues/2256
// Only use this when you are calling fmt.Fprintln, not log.Debug, etc.
func (c *Context) SetSensitiveValues(vals []string) {
	if len(vals) > 0 {
		out := NewCensoredWriter(c.Out)
		out.SetSensitiveValues(vals)
		c.Out = out

		err := NewCensoredWriter(c.Err)
		err.SetSensitiveValues(vals)
		c.Err = err
	}
}
