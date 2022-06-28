package storage

import (
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestBundleDocument_MarshalJSON(t *testing.T) {
	// Validate that we can marshal and unmarshal a bundle to an escaped json string
	b1 := BundleDocument(exampleBundle)

	data, err := json.Marshal(b1)
	require.NoError(t, err, "MarshalJSON failed")

	test.CompareGoldenFile(t, "testdata/marshalled_bundle.txt", string(data))

	var b2 BundleDocument
	err = json.Unmarshal(data, &b2)
	require.NoError(t, err, "UnmarshalJSON failed")
	require.Equal(t, b1, b2, "The bundle didn't survive the round trip")
}
