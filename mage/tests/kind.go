package tests

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"get.porter.sh/porter/mage/tools"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/pkg"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const (
	// Name of the KIND cluster used for testing
	DefaultKindClusterName = "porter"

	// Relative location of the KUBECONFIG for the test cluster
	Kubeconfig = "kind.config"

	// Container name of the local registry
	DefaultRegistryName = "registry"
)

var (
	must = shx.CommandBuilder{StopOnError: true}
)

// Ensure that the test KIND cluster is up.
func EnsureTestCluster() {
	mg.Deps(EnsureKubectl)

	if !useCluster() {
		CreateTestCluster()
	}
}

// get the config of the current kind cluster, if available
func getClusterConfig() (kubeconfig string, ok bool) {
	contents, err := shx.OutputE("kind", "get", "kubeconfig", "--name", getKindClusterName())
	return contents, err == nil
}

// setup environment to use the current kind cluster, if available
func useCluster() bool {
	contents, ok := getClusterConfig()
	if ok {
		log.Println("Reusing existing kind cluster")

		userKubeConfig, _ := filepath.Abs(os.Getenv("KUBECONFIG"))
		currentKubeConfig := filepath.Join(pwd(), Kubeconfig)
		if userKubeConfig != currentKubeConfig {
			fmt.Printf("ATTENTION! You should set your KUBECONFIG to match the cluster used by this project\n\n\texport KUBECONFIG=%s\n\n", currentKubeConfig)
		}
		os.Setenv("KUBECONFIG", currentKubeConfig)

		err := ioutil.WriteFile(Kubeconfig, []byte(contents), 0644)
		mgx.Must(errors.Wrapf(err, "error writing %s", Kubeconfig))
		return true
	}

	return false
}

func setClusterNamespace(name string) {
	must.RunE("kubectl", "config", "set-context", "--current", "--namespace", name)
}

// Create a KIND cluster named porter.
func CreateTestCluster() {
	mg.Deps(tools.EnsureKind, RestartDockerRegistry)

	// Determine host ip to populate kind config api server details
	// https://kind.sigs.k8s.io/docs/user/configuration/#api-server
	addrs, err := net.InterfaceAddrs()
	mgx.Must(errors.Wrap(err, "could not get a list of network interfaces"))

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

	os.Setenv("KUBECONFIG", filepath.Join(pwd(), Kubeconfig))
	kindCfgPath := "mage/tests/kind.config.yaml"
	kindCfg, err := ioutil.ReadFile(kindCfgPath)
	mgx.Must(errors.Wrapf(err, "error reading %s", kindCfgPath))

	kindCfgTmpl, err := template.New("kind.config.yaml").Parse(string(kindCfg))
	mgx.Must(errors.Wrapf(err, "error parsing EnsureKind config template %s", kindCfgPath))

	var kindCfgContents bytes.Buffer
	kindCfgData := struct {
		Address string
	}{
		Address: ipAddress,
	}
	err = kindCfgTmpl.Execute(&kindCfgContents, kindCfgData)
	err = ioutil.WriteFile("kind.config.yaml", kindCfgContents.Bytes(), 0644)
	mgx.Must(errors.Wrap(err, "could not write kind config file"))
	defer os.Remove("kind.config.yaml")

	must.Run("kind", "create", "cluster", "--name", getKindClusterName(), "--config", "kind.config.yaml")

	// Connect the kind and registry containers on the same network
	must.Run("docker", "network", "connect", "kind", getRegistryName())

	// Document the local registry
	kubectl("apply", "-f", "mage/tests/local-registry.yaml").Run()
}

// Delete the KIND cluster named porter.
func DeleteTestCluster() {
	mg.Deps(tools.EnsureKind)

	must.RunE("kind", "delete", "cluster", "--name", getKindClusterName())

	if isOnDockerNetwork(getRegistryName(), "kind") {
		must.RunE("docker", "network", "disconnect", "kind", getRegistryName())
	}
}

func kubectl(args ...string) shx.PreparedCommand {
	kubeconfig := fmt.Sprintf("KUBECONFIG=%s", os.Getenv("KUBECONFIG"))
	return must.Command("kubectl", args...).Env(kubeconfig)
}

func isOnDockerNetwork(container string, network string) bool {
	networkId, _ := shx.OutputE("docker", "network", "inspect", network, "-f", "{{.Id}}")
	networks, _ := shx.OutputE("docker", "inspect", container, "-f", "{{json .NetworkSettings.Networks}}")
	return strings.Contains(networks, networkId)
}

// Ensure the docker daemon is started and ready to accept connections.
func StartDocker() error {
	switch runtime.GOOS {
	case "windows":
		err := shx.RunS("powershell", "-c", "Get-Process 'Docker Desktop'")
		if err != nil {
			fmt.Println("Starting Docker Desktop")
			cmd := shx.Command(`C:\Program Files\Docker\Docker\Docker Desktop.exe`)
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

// Start a Docker registry to use with the tests.
func StartDockerRegistry() error {
	mg.Deps(StartDocker)
	if isContainerRunning(getRegistryName()) {
		return nil
	}

	err := RemoveContainer(getRegistryName())
	if err != nil {
		return err
	}

	fmt.Println("Starting local docker registry")
	return shx.RunE("docker", "run", "-d", "-p", "5000:5000", "--name", getRegistryName(), "registry:2")
}

// Stop the Docker registry used by the tests.
func StopDockerRegistry() error {
	if containerExists(getRegistryName()) {
		fmt.Println("Stopping local docker registry")
		return RemoveContainer(getRegistryName())
	}
	return nil
}

func RestartDockerRegistry() error {
	if err := StopDockerRegistry(); err != nil {
		return err
	}
	return StartDockerRegistry()
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

// Remove the specified container, if it is present.
func RemoveContainer(name string) error {
	stderr := bytes.Buffer{}
	_, _, err := shx.Command("docker", "rm", "-f", name).Stderr(&stderr).Stdout(nil).Exec()
	// Gracefully handle the container already being gone
	if err != nil && !strings.Contains(stderr.String(), "No such container") {
		return err
	}
	return nil
}

// Ensure kubectl is installed.
func EnsureKubectl() {
	if ok, _ := pkg.IsCommandAvailable("kubectl", ""); ok {
		return
	}

	versionURL := "https://storage.googleapis.com/kubernetes-release/release/stable.txt"
	versionResp, err := http.Get(versionURL)
	mgx.Must(errors.Wrapf(err, "unable to determine the latest version of kubectl"))

	if versionResp.StatusCode > 299 {
		mgx.Must(errors.Errorf("GET %s (%d): %s", versionURL, versionResp.StatusCode, versionResp.Status))
	}
	defer versionResp.Body.Close()

	kubectlVersion, err := ioutil.ReadAll(versionResp.Body)
	mgx.Must(errors.Wrapf(err, "error reading response from %s", versionURL))

	kindURL := "https://storage.googleapis.com/kubernetes-release/release/{{.VERSION}}/bin/{{.GOOS}}/{{.GOARCH}}/kubectl{{.EXT}}"
	mgx.Must(pkg.DownloadToGopathBin(kindURL, "kubectl", string(kubectlVersion)))
}

func pwd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(errors.Wrap(err, "pwd failed"))
	}
	return wd
}

func getKindClusterName() string {
	if name, ok := os.LookupEnv("KIND_NAME"); ok {
		return name
	}
	return DefaultKindClusterName
}

func getRegistryName() string {
	if name, ok := os.LookupEnv("REGISTRY_NAME"); ok {
		return name
	}
	return DefaultRegistryName
}
