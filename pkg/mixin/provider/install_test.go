package mixinprovider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_Install(t *testing.T) {
	// serve out a fake mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a mixxin\n")
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	c.SetupPorterHome()
	p := NewFileSystem(c.Config)

	opts := mixin.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	m, err := p.Install(opts)

	require.NoError(t, err)
	assert.Equal(t, "mixxin", m.Name)
	assert.Equal(t, "/root/.porter/mixins/mixxin", m.Dir)
	assert.Equal(t, "/root/.porter/mixins/mixxin/mixxin", m.ClientPath)
}
