package helm

import (
	"bytes"
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

// sad hack: not sure how to make a common test main for all my subpackages
func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}

func TestMixin_Install(t *testing.T) {
	os.Setenv(test.ExpectedCommandEnv, "helm install --name MYRELEASE MYCHART")
	defer os.Unsetenv(test.ExpectedCommandEnv)

	args := InstallArguments{
		Name:  "MYRELEASE",
		Chart: "MYCHART",
	}
	b, _ := yaml.Marshal(args)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	err := h.Install()

	require.NoError(t, err)
}
