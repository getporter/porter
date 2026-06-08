//go:build !windows

package client

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigureGracefulShutdown_ForwardsSIGTERM verifies that when the context
// is cancelled, the child process receives SIGTERM instead of SIGKILL, giving
// it time to perform cleanup (e.g. terraform flushing its state file) before
// the process is force-killed after the WaitDelay.
func TestConfigureGracefulShutdown_ForwardsSIGTERM(t *testing.T) {
	sentinelFile := filepath.Join(t.TempDir(), "sigterm-received")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// The child installs a SIGTERM trap, writes the sentinel, and exits cleanly.
	// The while-loop runs so the shell stays alive long enough for us to cancel.
	script := `trap 'touch "` + sentinelFile + `"; exit 0' TERM
while true; do sleep 0.05; done`

	cmd := exec.CommandContext(ctx, "sh", "-c", script)
	configureGracefulShutdown(cmd)

	require.NoError(t, cmd.Start())

	// Allow the shell to install its trap before cancelling.
	time.Sleep(150 * time.Millisecond)

	cancel()
	_ = cmd.Wait()

	_, statErr := os.Stat(sentinelFile)
	assert.NoError(t, statErr, "sentinel file must exist: the child process must receive SIGTERM, not SIGKILL")
}
