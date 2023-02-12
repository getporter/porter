package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobStatus_IsFailed(t *testing.T) {
	s := JobStatus{}
	assert.False(t, s.IsFailed(), "IsFailed should be false when the status is not failed")
	s.Status = "failed"
	assert.True(t, s.IsFailed(), "IsFailed should be true when the status is failed")
}

func TestJobStatus_IsCompleted(t *testing.T) {
	s := JobStatus{}
	assert.False(t, s.IsSucceeded(), "IsSucceeded should be false when the status is not succeeded")
	s.Status = "succeeded"
	assert.True(t, s.IsSucceeded(), "IsSucceeded should be true when the status is succeeded")
}
