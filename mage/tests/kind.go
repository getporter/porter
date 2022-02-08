package tests

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"get.porter.sh/porter/mage/docker"
	"get.porter.sh/porter/mage/tools"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/pkg"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
	"github.com/pkg/errors"
)

const (
	// Name of the KIND cluster used for testing
	DefaultKindClusterName = "porter"

	// Relative location of the KUBECONFIG for the test cluster
	Kubeconfig = "kind.config"
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
	mgx.Must(docker.StartDockerRegistry())
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

		err := ioutil.WriteFile(Kubeconfig, []byte(contents), 0600)
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
	mg.Deps(tools.EnsureKind, docker.RestartDockerRegistry)

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
	err = ioutil.WriteFile("kind.config.yaml", kindCfgContents.Bytes(), 0600)
	mgx.Must(errors.Wrap(err, "could not write kind config file"))
	defer os.Remove("kind.config.yaml")

	must.Run("kind", "create", "cluster", "--name", getKindClusterName(), "--config", "kind.config.yaml")

	// Document the local registry
	kubectl("apply", "-f", "mage/tests/local-registry.yaml").Run()
}

// Delete the KIND cluster named porter.
func DeleteTestCluster() {
	mg.Deps(tools.EnsureKind)

	must.RunE("kind", "delete", "cluster", "--name", getKindClusterName())
}

func kubectl(args ...string) shx.PreparedCommand {
	kubeconfig := fmt.Sprintf("KUBECONFIG=%s", os.Getenv("KUBECONFIG"))
	return must.Command("kubectl", args...).Env(kubeconfig)
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
