package claims

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallation_String(t *testing.T) {
	i := Installation{Name: "mybun"}
	assert.Equal(t, "/mybun", i.String())

	i.Namespace = "dev"
	assert.Equal(t, "dev/mybun", i.String())
}
