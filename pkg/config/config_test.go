package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GetHomeDir(t *testing.T) {
	// Do not run in parallel, relies on real environment variables

	t.Run("PORTER_HOME set", func(t *testing.T) {
		porterhome, _ := filepath.Abs("/home/myuser/.porter")
		t.Log("PORTER_HOME=", porterhome)
		c := NewTestConfig(t)
		c.porterHome = ""
		c.Setenv("PORTER_HOME", porterhome)

		assert.Equal(t, porterhome, c.GetHomeDir())
	})

	t.Run("HOME set", func(t *testing.T) {
		// Set the real env var because we are using go's implementation to find the user's home directory
		homeVar := "HOME"
		if runtime.GOOS == "windows" {
			homeVar = "USERPROFILE"
		}
		origHome := os.Getenv(homeVar)
		home, _ := filepath.Abs("/home/myuser")
		t.Logf("%s=%s", homeVar, home)
		os.Setenv(homeVar, home)
		defer os.Unsetenv(origHome)

		c := NewTestConfig(t)
		c.porterHome = ""
		c.Unsetenv("PORTER_HOME")

		assert.Equal(t, filepath.Join(home, ".porter"), c.GetHomeDir())
	})

	t.Run("unfindable", func(t *testing.T) {
		// Set the real env var because we are using go's implementation to find the user's home directory
		homeVar := "HOME"
		if runtime.GOOS == "windows" {
			homeVar = "USERPROFILE"
		}
		origHome := os.Getenv(homeVar)
		os.Unsetenv(homeVar)
		defer os.Unsetenv(origHome)

		c := NewTestConfig(t)
		c.porterHome = ""
		c.Unsetenv("PORTER_HOME")

		// Set the current working directory explicitly
		pwd, _ := filepath.Abs("/tmp")
		c.Chdir(pwd)

		// Home defaults to pwd when it can't be deduced otherwise
		gotHome := c.GetHomeDir()
		assert.Equal(t, pwd, gotHome)
	})
}
