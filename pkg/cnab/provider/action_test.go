package cnabprovider

import (
	"io/ioutil"
	"testing"

	"github.com/deislabs/cnab-go/driver"
	"github.com/deislabs/porter/pkg/config"
	instancestorageprovider "github.com/deislabs/porter/pkg/instance-storage/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddReloccation(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/relocation-mapping.json")
	require.NoError(t, err)

	c := config.NewTestConfig(t)
	instanceStorage := instancestorageprovider.NewPluginDelegator(c.Config)
	d := NewRuntime(c.Config, instanceStorage)

	args := ActionArguments{
		RelocationMapping: "/cnab/app/relocation-mapping.json",
	}

	c.TestContext.AddTestFile("testdata/relocation-mapping.json", "/cnab/app/relocation-mapping.json")

	opConf := d.AddRelocation(args)

	op := &driver.Operation{
		Files: make(map[string]string),
	}
	err = opConf(op)
	assert.NoError(t, err)

	mapping, ok := op.Files["/cnab/app/relocation-mapping.json"]
	assert.True(t, ok)
	assert.Equal(t, string(data), mapping)

}
