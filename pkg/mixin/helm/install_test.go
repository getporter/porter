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
	os.Setenv(test.ExpectedCommandEnv, `helm install --name MYRELEASE MYCHART --namespace MYNAMESPACE --version 1.0.0 --set "foo=bar"`)
	defer os.Unsetenv(test.ExpectedCommandEnv)

	args := InstallArguments{
		Namespace: "MYNAMESPACE",
		Name:      "MYRELEASE",
		Chart:     "MYCHART",
		Version:   "1.0.0",
		Set: map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}
	b, _ := yaml.Marshal(args)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	err := h.Install()

	require.NoError(t, err)
}
