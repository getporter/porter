package builder

import (
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/stretchr/testify/require"
)

var _ ExitError = TestExitError{}

type TestExitError struct {
	exitCode int
}

func (t TestExitError) Error() string {
	return "an error occurred"
}

func (t TestExitError) ExitCode() int {
	return t.exitCode
}

func TestIgnoreErrorHandler_HandleError(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{}
		err := h.HandleError(cxt.Context, nil, "", "")
		require.NoError(t, err)
	})

	t.Run("error - passthrough", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{}
		origErr := &TestExitError{1}
		err := h.HandleError(cxt.Context, origErr, "", "")
		require.Same(t, origErr, err)
	})

	t.Run("all", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{All: true}
		err := h.HandleError(cxt.Context, TestExitError{1}, "", "")
		require.NoError(t, err)
	})

	t.Run("allowed exit code", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{ExitCodes: []int{2, 1, 4}}
		err := h.HandleError(cxt.Context, TestExitError{1}, "", "")
		require.NoError(t, err)
	})

	t.Run("disallowed exit code", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{ExitCodes: []int{2, 1, 4}}
		origErr := &TestExitError{10}
		err := h.HandleError(cxt.Context, origErr, "", "")
		require.Same(t, origErr, err, "The original error should be preserved")
	})

	t.Run("stderr contains", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{Output: IgnoreErrorWithOutput{Contains: []string{"already exists"}}}
		origErr := &TestExitError{10}
		err := h.HandleError(cxt.Context, origErr, "", "The specified thing already exists")
		require.NoError(t, err)
	})

	t.Run("stderr does not contain", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{Output: IgnoreErrorWithOutput{Contains: []string{"already exists"}}}
		origErr := &TestExitError{10}
		err := h.HandleError(cxt.Context, origErr, "", "Something went wrong")
		require.Same(t, origErr, err)
	})

	t.Run("stderr matches regex", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{Output: IgnoreErrorWithOutput{Regex: []string{"(exists|EXISTS)"}}}
		origErr := &TestExitError{10}
		err := h.HandleError(cxt.Context, origErr, "", "something EXISTS")
		require.NoError(t, err)
	})

	t.Run("stderr does not match regex", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		h := IgnoreErrorHandler{Output: IgnoreErrorWithOutput{Regex: []string{"(exists|EXISTS)"}}}
		origErr := &TestExitError{10}
		err := h.HandleError(cxt.Context, origErr, "", "something mumble mumble")
		require.Same(t, origErr, err)
	})
}
