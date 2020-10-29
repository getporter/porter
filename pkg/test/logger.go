package test

import "testing"

// Logger helps capture output in a test while still showing it in the console
type Logger struct {
	T *testing.T
}

func (l Logger) Write(p []byte) (n int, err error) {
	l.T.Log(string(p))
	return len(p), nil
}
