// +build mage

// This is a magefile, and is a "makefile for go".
// See https://magefile.org/
package main

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/carolynvs/magex/pkg"
	"github.com/carolynvs/magex/shx"
	"github.com/carolynvs/magex/xplat"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

const registryContainer = "registry"

// Ensure Mage is installed and on the PATH.
func EnsureMage() error {
	return pkg.EnsureMage("")
}

// ConfigureAgent sets up an Azure DevOps agent with Mage and ensures
// that GOPATH/bin is in PATH.
func ConfigureAgent() error {
	err := EnsureMage()
	if err != nil {
		return err
	}

	// Instruct Azure DevOps to add GOPATH/bin to PATH
	gobin := xplat.FilePathJoin(xplat.GOPATH(), "bin")
	err = os.MkdirAll(gobin, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not mkdir -p %s", gobin)
	}
	fmt.Println("##vso[task.prependpath]/home/vsts/go/bin/")
	return nil
}

// Run end-to-end (e2e) tests
func TestE2E() error {
	mg.Deps(StartDocker, startLocalDockerRegistry)
	defer stopLocalDockerRegistry()

	// Only do verbose output of tests when called with `mage -v TestE2E`
	v := ""
	if mg.Verbose() {
		v = "-v"
	}

	return sh.RunV("go", shx.CollapseArgs("test", "-tags", "e2e", v, "./tests/e2e/...")...)
}

// Copy the cross-compiled binaries from xbuild into bin.
func UseXBuildBinaries() error {
	pwd, _ := os.Getwd()
	goos := build.Default.GOOS
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	copies := map[string]string{
		"bin/latest/porter-$GOOS-amd64$EXT":           "bin/porter$EXT",
		"bin/latest/porter-linux-amd64":               "bin/runtimes/porter-runtime",
		"bin/mixins/exec/latest/exec-$GOOS-amd64$EXT": "bin/mixins/exec/exec$EXT",
		"bin/mixins/exec/latest/exec-linux-amd64":     "bin/mixins/exec/runtimes/exec-runtime",
	}

	r := strings.NewReplacer("$GOOS", goos, "$EXT", ext, "$PWD", pwd)
	for src, dest := range copies {
		src = r.Replace(src)
		dest = r.Replace(dest)
		log.Printf("Copying %s to %s", src, dest)

		destDir := filepath.Dir(dest)
		os.MkdirAll(destDir, 0755)

		err := sh.Copy(dest, src)
		if err != nil {
			return err
		}
	}

	return SetBinExecutable()
}

// Run `chmod +x -R bin`.
func SetBinExecutable() error {
	err := chmodRecursive("bin", 0755)
	return errors.Wrap(err, "could not set +x on the test bin")
}

func chmodRecursive(name string, mode os.FileMode) error {
	return filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		log.Println("chmod +x ", path)
		return os.Chmod(path, mode)
	})
}

// Ensure the docker daemon is started and ready to accept connections.
func StartDocker() error {
	switch runtime.GOOS {
	case "windows":
		_, err := shx.OutputS("powershell", "-c", "Get-Process 'Docker Desktop'")
		if err != nil {
			log.Println("Starting Docker Desktop")
			ran, err := sh.Exec(nil, nil, nil, `C:\Program Files\Docker\Docker\Docker Desktop.exe`)
			if !ran {
				return errors.Wrapf(err, "could not start Docker Desktop")
			}
		}
	}

	log.Print("Waiting for the docker service to be ready")
	ready := false
	for count := 0; count < 60; count++ {
		err := shx.RunS("docker", "ps")
		if !sh.CmdRan(err) {
			return errors.Wrap(err, "could not run docker")
		}
		if err == nil {
			ready = true
			break
		}
		log.Print(".")
		time.Sleep(time.Second)
	}
	log.Println()

	if !ready {
		return errors.New("a timeout was reached waiting for the docker service to become unavailable")
	}

	log.Println("Docker service is ready!")
	return nil
}

func startLocalDockerRegistry() error {
	if !isContainerRunning(registryContainer) {
		log.Println("Starting local docker registry")
		return shx.RunE("docker", "run", "-d", "-p", "5000:5000", "--name", registryContainer, "registry:2")
	}
	return nil
}

func stopLocalDockerRegistry() error {
	if containerExists(registryContainer) {
		log.Println("Stopping local docker registry")
		return removeContainer(registryContainer)
	}
	return nil
}

func isContainerRunning(name string) bool {
	out, _ := shx.OutputS("docker", "container", "inspect", "-f", "{{.State.Running}}", name)
	running, _ := strconv.ParseBool(out)
	return running
}

func containerExists(name string) bool {
	err := shx.RunS("docker", "inspect", name)
	return err == nil
}

func removeContainer(name string) error {
	return shx.RunE("docker", "rm", "-f", name)
}
