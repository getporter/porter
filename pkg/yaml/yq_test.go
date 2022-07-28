package yaml_test

import (
	"context"
	"errors"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/mikefarah/yq/v3/pkg/yqlib"
	"github.com/stretchr/testify/require"
)

func TestEditor_WalkNodes(t *testing.T) {
	pCtx := portercontext.NewTestContext(t)
	defer pCtx.Close()
	pCtx.AddTestFile("testdata/custom.yaml", config.Name)

	e := yaml.NewEditor(pCtx.Context)
	err := e.ReadFile(config.Name)
	require.NoError(t, err)

	testcases := []struct {
		name       string
		path       string
		totalNodes int
		wantErr    error
	}{
		{name: "success", path: "root.array.**.a", totalNodes: 2},
		{name: "no matching nodes found", path: "hello", totalNodes: 0},
		{name: "exit walking on first error", path: "root.array", totalNodes: 1, wantErr: errors.New("failed callback")},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var totalNodes int
			err = e.WalkNodes(context.Background(), tc.path, func(ctx context.Context, nc *yqlib.NodeContext) error {
				totalNodes += 1
				return tc.wantErr
			})

			require.Equal(t, tc.totalNodes, totalNodes)
			require.Equal(t, tc.wantErr, err)
		})
	}

}
