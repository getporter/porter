package porter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPorter_printOutputsTable(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

	want := `----------------------------
  Name  Type    Value         
----------------------------
  bar   string  bar-value     
  foo   string  /path/to/foo  
`

	outputs := []DisplayOutput{
		{Name: "bar", Type: "string", DisplayValue: "bar-value"},
		{Name: "foo", Type: "string", DisplayValue: "/path/to/foo"},
	}
	err := p.printOutputsTable(outputs)
	require.NoError(t, err)

	got := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, want, got)
}
