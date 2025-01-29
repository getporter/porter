package agent

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg"
)

// allow the tests to capture output
var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

// The porter agent wraps the porter cli,
// handling copying config files from a mounted
// volume into PORTER_HOME
// Returns any errors and if the porter command was executed
func Execute(porterCommand []string, porterHome string, porterConfig string) (error, bool) {
	porter := porterHome + "/porter"

	// Copy config files into PORTER_HOME
	err := filepath.Walk(porterConfig, func(path string, info fs.FileInfo, err error) error {
		// if Walk sends a non nil err, then return it back
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Determine the relative path of the file we are copying
		relPath, err := filepath.Rel(porterConfig, path)
		if err != nil {
			return err
		}

		// Skip hidden files, these are injected by k8s when the config volume is mounted
		if strings.HasPrefix(relPath, ".") {
			return nil
		}

		// If the files are symlinks then resolve them
		// /porter-config
		//    - config.toml (symlink to the file in ..data)
		//    - ..data/config.toml
		resolvedPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}

		resolvedInfo, err := os.Stat(resolvedPath)
		if err != nil {
			return err
		}

		return copyConfig(relPath, resolvedPath, resolvedInfo, porterHome)
	})
	if err != nil {
		return err, false
	}

	// Remind everyone the version of Porter we are using
	fmt.Fprintf(Stderr, "porter version\n")
	cmd := exec.Command(porter, "version")
	cmd.Stdout = Stderr // send all non-bundle output to stderr
	cmd.Stderr = Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("porter version check failed: %w", err), false
	}

	// Run the specified porter command
	fmt.Fprintf(Stderr, "porter %s\n", strings.Join(porterCommand, " "))
	cmd = exec.Command(porter, porterCommand...)
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return err, false
	}
	return cmd.Wait(), true
}

func copyConfig(relPath string, configFile string, fi os.FileInfo, porterHome string) error {
	destFile := filepath.Join(porterHome, relPath)
	fmt.Fprintln(Stderr, "Loading configuration", relPath, "into", destFile)
	src, err := os.OpenFile(configFile, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(destFile), pkg.FileModeDirectory); err != nil {
		return err
	}
	dest, err := os.OpenFile(destFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fi.Mode())
	if err != nil {
		return err
	}
	defer dest.Close()

	// Just copy the file
	_, err = io.Copy(dest, src)
	return err
}
