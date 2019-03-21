package ohai

import (
	"fmt"
	"io"
)

// Ohai displays an informative message.
func Ohai(w io.Writer, a ...interface{}) (int, error) {
	return Ohaif(w, "%s", a...)
}

// Ohaif displays an informative message.
func Ohaif(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, fmt.Sprintf("==> %s", format), a...)
}

// Ohailn displays an informative message.
func Ohailn(w io.Writer, a ...interface{}) (int, error) {
	return Ohaif(w, "%s\n", a...)
}

// Fohai displays an informative message.
func Fohai(w io.Writer, a ...interface{}) (int, error) {
	return Fohaif(w, "%s", a...)
}

// Fohaif displays an informative message.
func Fohaif(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, fmt.Sprintf("==> %s", format), a...)
}

// Fohailn displays an informative message.
func Fohailn(w io.Writer, a ...interface{}) (int, error) {
	return Fohaif(w, "%s\n", a...)
}

// Success displays a success message.
func Success(w io.Writer, a ...interface{}) (int, error) {
	return Successf(w, "%s", a...)
}

// Successf displays a success message.
func Successf(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, fmt.Sprintf("✓✓✓ %s", format), a...)
}

// Successln displays a success message.
func Successln(w io.Writer, a ...interface{}) (int, error) {
	return Successf(w, "%s\n", a...)
}

// Fsuccess displays an informative message.
func Fsuccess(w io.Writer, a ...interface{}) (int, error) {
	return Fsuccessf(w, "%s", a...)
}

// Fsuccessf displays an informative message.
func Fsuccessf(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, fmt.Sprintf("==> %s", format), a...)
}

// Fsuccessln displays an informative message.
func Fsuccessln(w io.Writer, a ...interface{}) (int, error) {
	return Fsuccessf(w, "%s\n", a...)
}

// Warning displays a warning message.
func Warning(w io.Writer, a ...interface{}) (int, error) {
	return Warningf(w, "%s", a...)
}

// Warningf displays a warning message.
func Warningf(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, fmt.Sprintf("!!! %s", format), a...)
}

// Warningln displays a warning message.
func Warningln(w io.Writer, a ...interface{}) (int, error) {
	return Warningf(w, "%s\n", a...)
}

// Fwarning displays an informative message.
func Fwarning(w io.Writer, a ...interface{}) (int, error) {
	return Fwarningf(w, "%s", a...)
}

// Fwarningf displays an informative message.
func Fwarningf(w io.Writer, format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(w, fmt.Sprintf("==> %s", format), a...)
}

// Fwarningln displays an informative message.
func Fwarningln(w io.Writer, a ...interface{}) (int, error) {
	return Fwarningf(w, "%s\n", a...)
}
