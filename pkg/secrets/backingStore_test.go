package secrets

import (
	"testing"

	inmemory "get.porter.sh/porter/pkg/secrets/in-memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackingStore_Resolve(t *testing.T) {
	const TestItemType = "test-items"

	testcases := []struct {
		name      string
		autoclose bool
	}{
		{name: "Default AutoClose Connections", autoclose: true},
		{name: "Self Managed Connections", autoclose: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := inmemory.NewStore()
			s.Secrets[TestItemType] = map[string]string{"key1": "value1"}
			bs := NewSecretStore(s)
			bs.AutoClose = tc.autoclose

			val, err := bs.Resolve(TestItemType, "key1")
			require.NoError(t, err, "expected Resolve to succeed")
			assert.Equal(t, "value1", string(val), "Resolve returned the wrong data")

			connectCount, err := s.GetConnectCount()
			require.NoError(t, err, "GetConnectCount failed")
			assert.Equal(t, 1, connectCount, "Connect should have been called once")

			closeCount, err := s.GetCloseCount()
			require.NoError(t, err, "GetCloseCount failed")
			if tc.autoclose {
				assert.Equal(t, 1, closeCount, "Close should have been automatically called after Read")
			} else {
				assert.Equal(t, 0, closeCount, "Close should not be automatically called")
			}
		})
	}
}
