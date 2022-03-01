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
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// allow the tests to capture output
var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

// The porter agent wraps the porter cli,
// handling copying config files from a mounted
// volume into PORTER_HOME
// Returns any errors and if the porter command was executed
func Execute(porterCommand []string, porterHome string, porterConfig string) (error, bool) {
	porter := porterHome + "/porter"

	// Copy config files into PORTER_HOME
	err := filepath.Walk(porterConfig, func(path string, info fs.FileInfo, err error) error {
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
	fmt.Fprintf(stderr, "porter version\n")
	cmd := exec.Command(porter, "version")
	cmd.Stdout = stderr // send all non-bundle output to stderr
	cmd.Stderr = stderr
	if err = cmd.Run(); err != nil {
		return errors.Wrap(err, "porter version check failed"), false
	}

	// Run the specified porter command
	fmt.Fprintf(stderr, "porter %s\n", strings.Join(porterCommand, " "))
	cmd = exec.Command(porter, porterCommand...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return err, false
	}
	return cmd.Wait(), true
}

func copyConfig(relPath string, configFile string, fi os.FileInfo, porterHome string) error {
	destFile := filepath.Join(porterHome, relPath)
	fmt.Fprintln(stderr, "Loading configuration", relPath, "into", destFile)
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

	if !isExecutable(fi.Mode()) {
		// Copy the file and write out its content at the same time
		wg := errgroup.Group{}
		pr, pw := io.Pipe()
		tr := io.TeeReader(src, pw)

		// Copy the File
		wg.Go(func() error {
			defer pw.Close()

			_, err = io.Copy(dest, tr)
			return err
		})

		// Print out the contents of the transferred file only if it's not executable
		wg.Go(func() error {
			// read from the PipeReader to stdout
			_, err := io.Copy(stderr, pr)

			// Pad with whitespace so it's easier to see the file contents
			fmt.Fprintf(stderr, "\n\n")
			return err
		})

		return wg.Wait()
	}

	// Just copy the file if it's binary, don't print it out
	_, err = io.Copy(dest, src)
	return err
}

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
