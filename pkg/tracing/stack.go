package tracing

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
)

const (
	modulePrefix = "get.porter.sh/porter/"

	// maxStackDepth is the maximum number of frames captured for an error.
	maxStackDepth = 32
)

// loggerType is the logger that captures stack traces. Its functions are
// removed from the stack trace.
var loggerType = reflect.TypeFor[traceLogger]()

// moduleRoot is the directory that contained the porter source code when the
// binary was built, with a trailing slash. It is used to print file paths
// relative to the root of the repository.
var moduleRoot = findModuleRoot()

func findModuleRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	// Remove the directory of this package from the directory of this file
	pkgDir := strings.TrimPrefix(loggerType.PkgPath(), modulePrefix)
	root, ok := strings.CutSuffix(path.Dir(file), pkgDir)
	if !ok {
		return ""
	}
	return root
}

// stackError decorates an error with the stack trace of where it was first
// recorded by a TraceLogger.
type stackError struct {
	err error
	pcs []uintptr
}

func (e *stackError) Error() string {
	return e.err.Error()
}

func (e *stackError) Unwrap() error {
	return e.err
}

// withStack attaches the stack trace of the caller to the error.
// When the error already has a stack trace it is returned as-is, so that the
// stack trace closest to the origin of the error is kept.
func withStack(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := errors.AsType[*stackError](err); ok {
		return err
	}

	pcs := make([]uintptr, maxStackDepth)
	// skip runtime.Callers and withStack
	n := runtime.Callers(2, pcs)
	return &stackError{err: err, pcs: pcs[:n]}
}

// frames returns the stack trace with the innermost function first, without
// the functions that don't help locate the source of the error.
func (e *stackError) frames() []runtime.Frame {
	if len(e.pcs) == 0 {
		return nil
	}

	var result []runtime.Frame
	frames := runtime.CallersFrames(e.pcs)
	for {
		frame, more := frames.Next()
		if frame.Function != "" && !isNoiseFrame(frame.Function) {
			result = append(result, frame)
		}
		if !more {
			break
		}
	}
	return result
}

// StackTrace returns a stack trace of where the error was first recorded by a
// TraceLogger, with the innermost function first. It is formatted to be read
// by a person on the console.
// Returns an empty string when the error does not have a stack trace.
func StackTrace(err error) string {
	stackErr, ok := errors.AsType[*stackError](err)
	if !ok {
		return ""
	}
	frames := stackErr.frames()
	if len(frames) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("Stack trace:")
	for _, frame := range frames {
		fmt.Fprintf(&b, "\n  at %s (%s:%d)",
			strings.TrimPrefix(frame.Function, modulePrefix),
			formatFile(frame.File), frame.Line)
	}
	return b.String()
}

// telemetryStackTrace returns the stack trace of the error in the format used
// by the go runtime, which is what the exception.stacktrace attribute of a
// span should contain.
// Returns an empty string when the error does not have a stack trace.
func telemetryStackTrace(err error) string {
	stackErr, ok := errors.AsType[*stackError](err)
	if !ok {
		return ""
	}

	var b strings.Builder
	for _, frame := range stackErr.frames() {
		fmt.Fprintf(&b, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
	}
	return b.String()
}

// isNoiseFrame returns if the function doesn't help locate the source of an
// error, such as the logger that captured the stack trace or the go runtime.
func isNoiseFrame(function string) bool {
	return strings.HasPrefix(function, "runtime.") ||
		strings.HasPrefix(function, loggerType.PkgPath()+"."+loggerType.Name()+".") ||
		strings.HasPrefix(function, loggerType.PkgPath()+".(*"+loggerType.Name()+").")
}

// formatFile shortens the path to a source file. Files in porter are relative
// to the root of the repository, and dependencies start with their module path.
func formatFile(file string) string {
	if moduleRoot != "" {
		if rel, ok := strings.CutPrefix(file, moduleRoot); ok {
			return rel
		}
	}
	if _, rel, ok := strings.Cut(file, "/pkg/mod/"); ok {
		return rel
	}
	return file
}
