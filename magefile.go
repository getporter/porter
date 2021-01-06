// +build mage

// This is a magefile, and is a "makefile for go".
// See https://magefile.org/
package main

import (
	"context"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/carolynvs/magex/pkg"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

const (
	registryContainer = "registry"
	mixinsURL         = "https://cdn.porter.sh/mixins/"
	kindVeresion      = "v0.9.0"
	kindClusterName   = "porter"
)

// Ensure Mage is installed and on the PATH.
func EnsureMage() error {
	return pkg.EnsureMage("")
}

// Configure an Azure DevOps agent with Mage and ensures
// that GOPATH/bin is in PATH.
func ConfigureAgent() error {
	mg.SerialDeps(pkg.EnsureGopathBin, EnsureMage)
	return nil
}

// Install mixins used by tests and example bundles, if not already installed.
func GetMixins() error {
	mixinTag := os.Getenv("MIXIN_TAG")
	if mixinTag == "" {
		mixinTag = "canary"
	}

	mixins := []string{"helm", "arm", "terraform", "kubernetes"}
	var errG errgroup.Group
	for _, mixin := range mixins {
		mixinDir := filepath.Join("bin/mixins/", mixin)
		if _, err := os.Stat(mixinDir); err == nil {
			log.Println("Mixin already installed into bin:", mixin)
			continue
		}

		mixin := mixin
		errG.Go(func() error {
			log.Println("Installing mixin:", mixin)
			mixinURL := mixinsURL + mixin
			_, _, err := porter("mixin", "install", mixin, "--version", mixinTag, "--url", mixinURL).Run()
			return err
		})
	}

	return errG.Wait()
}

// Run a porter command from the bin
func porter(args ...string) sh.PreparedCommand {
	porterPath := filepath.Join("bin", "porter")
	p := sh.Command(porterPath, args...)

	porterHome, _ := filepath.Abs("bin")
	p.Cmd.Env = []string{"PORTER_HOME=" + porterHome}

	return p
}

// Run integration tests. Set $USE_CURRENT_CLUSTER=true to
// use your current cluster context. Otherwise a new KIND
// cluster will be created just for the test run.
func TestIntegration() error {
	deps := []interface{}{startLocalDockerRegistry, GetMixins}
	if os.Getenv("USE_CURRENT_CLUSTER") != "true" {
		deps = append(deps, CreateKindCluster)
		defer DeleteKindCluster()
	}
	mg.Deps(deps...)
	defer stopLocalDockerRegistry()

	err := sh.RunWith(map[string]string{"CGO_ENABLED": "0"},
		"go", "build", "-o", "bin/testplugin"+pkgmgmt.FileExt, "./cmd/testplugin")
	if err != nil {
		return errors.Wrap(err, "could not build test plugin")
	}

	// Only do verbose output of tests when called with `mage -v TestE2E`
	v := ""
	if mg.Verbose() {
		v = "-v"
	}

	pwd, _ := os.Getwd()
	return sh.RunWithV(map[string]string{"PROJECT_ROOT": pwd, "CGO_ENABLED": "0"},
		"go", shx.CollapseArgs("test", "-timeout", "30m", "-tags=integration", v, "./...")...)
}

// Run end-to-end (e2e) tests.
func TestE2E() error {
	mg.Deps(startLocalDockerRegistry)
	defer stopLocalDockerRegistry()

	// Only do verbose output of tests when called with `mage -v TestE2E`
	v := ""
	if mg.Verbose() {
		v = "-v"
	}

	return sh.RunWithV(map[string]string{"CGO_ENABLED": "0"},
		"go", shx.CollapseArgs("test", "-tags=e2e", v, "./tests/e2e/...")...)
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
		err := shx.RunS("powershell", "-c", "Get-Process 'Docker Desktop'")
		if err != nil {
			fmt.Println("Starting Docker Desktop")
			cmd := sh.Command(`C:\Program Files\Docker\Docker\Docker Desktop.exe`)
			err := cmd.Cmd.Start()
			if err != nil {
				return errors.Wrapf(err, "could not start Docker Desktop")
			}
		}
	}

	ready, err := isDockerReady()
	if err != nil {
		return err
	}

	if ready {
		return nil
	}

	fmt.Println("Waiting for the docker service to be ready")
	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	for {
		select {
		case <-cxt.Done():
			return errors.New("a timeout was reached waiting for the docker service to become unavailable")
		default:
			// Wait and check again
			// Writing a dot on a single line so the CI logs show our progress, instead of a bunch of dots at the end
			fmt.Println(".")
			time.Sleep(time.Second)

			if ready, _ := isDockerReady(); ready {
				fmt.Println("Docker service is ready!")
				return nil
			}
		}
	}
}

func isDockerReady() (bool, error) {
	err := shx.RunS("docker", "ps")
	if !sh.CmdRan(err) {
		return false, errors.Wrap(err, "could not run docker")
	}

	return err == nil, nil
}

func startLocalDockerRegistry() error {
	mg.Deps(StartDocker)
	if isContainerRunning(registryContainer) {
		return nil
	}

	err := removeContainer(registryContainer)
	if err != nil {
		return err
	}

	fmt.Println("Starting local docker registry")
	return shx.RunE("docker", "run", "-d", "-p", "5000:5000", "--name", registryContainer, "registry:2")
}

func stopLocalDockerRegistry() error {
	if containerExists(registryContainer) {
		fmt.Println("Stopping local docker registry")
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
	stderr, err := shx.OutputE("docker", "rm", "-f", name)
	// Gracefully handle the container already being gone
	if err != nil && !strings.Contains(stderr, "No such container") {
		return err
	}
	return nil
}

// Create a KIND cluster (kind-porter).
func CreateKindCluster() error {
	mg.Deps(EnsureKind)

	// Determine host ip to populate kind config api server details
	// https://kind.sigs.k8s.io/docs/user/configuration/#api-server
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return errors.Wrap(err, "could not get a list of network interfaces")
	}

	var ipAddress string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("Current IP address : ", ipnet.IP.String())
				ipAddress = ipnet.IP.String()
				break
			}
		}
	}

	kindCfg := "kind.config.yaml"
	contents := fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: %s
`, ipAddress)
	err = ioutil.WriteFile(kindCfg, []byte(contents), 0644)
	if err != nil {
		return errors.Wrap(err, "could not write kind config file")
	}
	defer os.Remove(kindCfg)

	err = shx.RunE("kind", "create", "cluster", "--name", kindClusterName, "--config", kindCfg)
	return errors.Wrap(err, "could not create KIND cluster")
}

// Delete the KIND cluster (kind-porter).
func DeleteKindCluster() error {
	err := shx.RunE("kind", "delete", "cluster", "--name", kindClusterName)
	return errors.Wrap(err, "could not delete KIND cluster")
}

// Ensure that KIND is installed and on the PATH.
func EnsureKind() error {
	if ok, _ := pkg.IsCommandAvailable("kind", ""); ok {
		return nil
	}

	kindURL := fmt.Sprintf("https://github.com/kubernetes-sigs/kind/releases/download/%s/kind-%s-%s", kindVeresion, runtime.GOOS, runtime.GOARCH)
	err := pkg.DownloadToGopathBin(kindURL, "kind")
	if err != nil {
		return errors.Wrap(err, "could not download kind")
	}

	return nil
}
