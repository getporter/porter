package signals

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterruptSignals_IncludesOSInterrupt(t *testing.T) {
	sigs := InterruptSignals()
	assert.Contains(t, sigs, os.Interrupt, "InterruptSignals must include os.Interrupt on all platforms")
}
