package claims

import (
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
)

func TestRun_NewResultFrom(t *testing.T) {
	run := NewRun("dev", "mybuns")
	cnabResult := cnab.Result{
		ID:             "resultID",
		ClaimID:        "claimID",
		Created:        time.Now(),
		Message:        "message",
		Status:         "status",
		OutputMetadata: OutputMetadata{"myoutput": struct{}{}},
		Custom:         map[string]interface{}{"custom": true},
	}

	result := run.NewResultFrom(cnabResult)
	assert.Equal(t, cnabResult.ID, result.ID)
	assert.Equal(t, run.Namespace, result.Namespace)
	assert.Equal(t, run.Installation, result.Installation)
	assert.Equal(t, run.ID, result.RunID)
	assert.Equal(t, cnabResult.Created, result.Created)
	assert.Equal(t, cnabResult.Status, result.Status)
	assert.Equal(t, cnabResult.Message, result.Message)
	assert.Equal(t, cnabResult.OutputMetadata, result.OutputMetadata)
	assert.Equal(t, cnabResult.Custom, result.Custom)
}

func TestRun_ShouldRecord(t *testing.T) {
	t.Run("stateless, not modifies", func(t *testing.T) {
		b := bundle.Bundle{
			Actions: map[string]bundle.Action{
				"dry-run": {
					Modifies:  false,
					Stateless: true,
				},
			},
		}

		r := Run{Bundle: b, Action: "dry-run"}
		assert.False(t, r.ShouldRecord())
	})

	t.Run("stateful, not modifies", func(t *testing.T) {
		b := bundle.Bundle{
			Actions: map[string]bundle.Action{
				"audit": {
					Modifies:  false,
					Stateless: false,
				},
			},
		}

		r := Run{Bundle: b, Action: "audit"}
		assert.True(t, r.ShouldRecord())
	})

	t.Run("modifies", func(t *testing.T) {
		b := bundle.Bundle{
			Actions: map[string]bundle.Action{
				"editstuff": {
					Modifies:  true,
					Stateless: false,
				},
			},
		}

		r := Run{Bundle: b, Action: "editstuff"}
		assert.True(t, r.ShouldRecord())
	})

	t.Run("missing definition", func(t *testing.T) {
		b := bundle.Bundle{}

		r := Run{Bundle: b, Action: "missing"}
		assert.True(t, r.ShouldRecord())
	})

}
