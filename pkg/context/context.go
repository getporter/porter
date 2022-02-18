package context

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
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// MixinOutputsDir represents the directory where mixin output files are written/read
	MixinOutputsDir = "/cnab/app/porter/outputs"
)

type CommandBuilder func(name string, arg ...string) *exec.Cmd

type Context struct {
	Debug              bool
	DebugPlugins       bool
	verbose            bool
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

	// logLevel filters the messages written to the console and logfile
	logLevel zapcore.Level
	logFile  afero.File

	// indicates if log timestamps should be printed to the console
	timestampLogs bool

	// handles sending tracing data to an otel collector
	tracer trace.Tracer

	// handles send log data to the console/logfile
	logger *zap.Logger

	// cleans up resources associated with the tracer when porter completes
	traceCloser *sdktrace.TracerProvider

	// the service name sent to the otel collector when we send tracing data
	traceServiceName string
}

// New creates a new context in the specified directory.
func New() *Context {
	// Ignore any error getting the working directory and report errors
	// when we attempt to access files in the current directory. This
	// allows us to use the current directory as a default, and allow
	// tests to override it.
	pwd, _ := os.Getwd()

	c := &Context{
		environ:       getEnviron(),
		FileSystem:    aferox.NewAferox(pwd, afero.NewOsFs()),
		In:            os.Stdin,
		Out:           NewCensoredWriter(os.Stdout),
		Err:           NewCensoredWriter(os.Stderr),
		correlationId: cnabclaims.MustNewULID(), // not using cnab package because that creates a cycle
		timestampLogs: true,
	}

	c.ConfigureLogging(LogConfiguration{})
	c.defaultNewCommand()
	c.PlugInDebugContext = NewPluginDebugContext(c)

	return c
}

// StartRootSpan creates the root tracing span for the porter application.
// This should only be done once.
func (c *Context) StartRootSpan(ctx context.Context, op string, attrs ...attribute.KeyValue) (context.Context, tracing.TraceLogger) {
	childCtx, span := c.tracer.Start(ctx, op)
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
	LogToFile            bool
	LogDirectory         string
	LogLevel             zapcore.Level
	StructuredLogs       bool
	TelemetryEnabled     bool
	TelemetryEndpoint    string
	TelemetryProtocol    string
	TelemetryInsecure    bool
	TelemetryCertificate string
	TelemetryCompression string
	TelemetryTimeout     string
	TelemetryHeaders     map[string]string
}

// ConfigureLogging applies different configuration to our logging and tracing.
func (c *Context) ConfigureLogging(cfg LogConfiguration) {
	// Cleanup in case logging has been configured before
	c.logLevel = cfg.LogLevel

	encoding := c.makeLogEncoding()
	consoleLogger := c.makeConsoleLogger(encoding, cfg.StructuredLogs)

	// make a temporary logger that we can use until we've completely initialized the full logger
	tmpLog := zap.New(consoleLogger)

	var err error
	fileLogger := zapcore.NewNopCore()
	if cfg.LogToFile {
		fileLogger, err = c.configureFileLog(encoding, cfg.LogDirectory)
		if err != nil {
			tmpLog.Error(errors.Wrap(err, "could not configure a file logger").Error())
		} else {
			tmpLog.Debug("Writing logs to " + c.logFile.Name())
		}
	}
	tmpLog = zap.New(zapcore.NewTee(consoleLogger, fileLogger))

	if cfg.TelemetryEnabled {
		// Only initialize the tracer once per command
		if c.traceCloser == nil {
			err = c.configureTelemetry(tmpLog, cfg)
			if err != nil {
				tmpLog.Error(errors.Wrap(err, "could not configure a tracer").Error())
			}
		}
	} else {
		c.tracer = trace.NewNoopTracerProvider().Tracer("noop")
	}

	c.logger = tmpLog
}

func (c *Context) makeConsoleLogger(encoding zapcore.EncoderConfig, structuredLogs bool) zapcore.Core {
	stderr := c.Err
	if f, ok := stderr.(*os.File); ok {
		if isatty.IsTerminal(f.Fd()) {
			stderr = colorable.NewColorable(f)
			encoding.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		}
	}

	// if structured-logs feature isn't enabled, keep the logs looking like they do now, with just the message printed
	if !structuredLogs {
		encoding.TimeKey = ""
		encoding.LevelKey = ""
	}
	consoleEncoder := zapcore.NewConsoleEncoder(encoding)
	return zapcore.NewCore(consoleEncoder, zapcore.AddSync(stderr), c.logLevel)
}

func (c *Context) configureFileLog(encoding zapcore.EncoderConfig, dir string) (zapcore.Core, error) {
	if err := c.FileSystem.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	// Write the logs to a file
	logfile := filepath.Join(dir, c.correlationId+".json")
	if c.logFile == nil { // We may have already opened this logfile, and we are just changing the log level
		f, err := c.FileSystem.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return zapcore.NewNopCore(), errors.Wrapf(err, "could not start log file at %s", logfile)
		}
		c.logFile = f
	}

	// Split logs to the console and file
	fileEncoder := zapcore.NewJSONEncoder(encoding)
	return zapcore.NewCore(fileEncoder, zapcore.AddSync(c.logFile), c.logLevel), nil
}

func (c *Context) Close() error {
	c.closeLogger()
	if c.traceCloser != nil {
		c.traceCloser.Shutdown(context.TODO())
	}
	return nil
}

func (c *Context) closeLogger() {
	if c.logFile != nil {
		c.logFile.Close()
		c.logFile = nil
	}
}

func (c *Context) defaultNewCommand() {
	c.NewCommand = func(name string, arg ...string) *exec.Cmd {
		return c.Command(name, arg...)
	}
}

// Command creates a new exec.Cmd using the context's current directory.
func (c *Context) Command(name string, arg ...string) *exec.Cmd {
	cmd := &exec.Cmd{
		Dir:  c.Getwd(),
		Path: name,
		Args: append([]string{name}, arg...),
	}
	if filepath.Base(name) == name {
		if lp, ok := c.LookPath(name); ok {
			cmd.Path = lp
		}
	}
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

func (c *Context) SetVerbose(value bool) {
	c.verbose = value
}

func (c *Context) IsVerbose() bool {
	return c.Debug || c.verbose
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
			return errors.WithStack(err)
		}

		// Translate the path from the src to the final destination
		dest := filepath.Join(destDir, strings.TrimPrefix(path, stripPrefix))
		if dest == "" {
			return nil
		}

		if info.IsDir() {
			return errors.WithStack(c.FileSystem.MkdirAll(dest, info.Mode()))
		}

		return c.CopyFile(path, dest)
	})
}

func (c *Context) CopyFile(src, dest string) error {
	info, err := c.FileSystem.Stat(src)
	if err != nil {
		return errors.WithStack(err)
	}

	data, err := c.FileSystem.ReadFile(src)
	if err != nil {
		return errors.WithStack(err)
	}

	err = c.FileSystem.WriteFile(dest, data, info.Mode())
	return errors.WithStack(err)
}

// WriteMixinOutputToFile writes the provided bytes (representing a mixin output)
// to a file named by the provided filename in Porter's mixin outputs directory
func (c *Context) WriteMixinOutputToFile(filename string, bytes []byte) error {
	exists, err := c.FileSystem.DirExists(MixinOutputsDir)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.FileSystem.MkdirAll(MixinOutputsDir, 0700); err != nil {
			return errors.Wrap(err, "couldn't make output directory")
		}
	}

	return c.FileSystem.WriteFile(filepath.Join(MixinOutputsDir, filename), bytes, 0600)
}

// SetSensitiveValues sets the sensitive values needing masking on output/err streams
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

// UserAgent returns a string that can be used as a user agent for porter.
func (c *Context) UserAgent() string {
	product := "porter"

	if pkg.Commit == "" && pkg.Version == "" {
		return product
	}

	v := pkg.Version
	if len(v) == 0 {
		v = pkg.Commit
	}

	return product + "/" + v
}
