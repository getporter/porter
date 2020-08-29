//go:generate packr2

package kubernetes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

const (
	defaultKubernetesClientVersion string = "v1.15.5"
)

type Mixin struct {
	*context.Context
	schemas                 *packr.Box
	KubernetesClientVersion string
}

func New() *Mixin {
	return &Mixin{
		Context:                 context.New(),
		schemas:                 NewSchemaBox(),
		KubernetesClientVersion: defaultKubernetesClientVersion,
	}
}

func NewSchemaBox() *packr.Box {
	return packr.New("get.porter.sh/porter/pkg/kubernetes/schema", "./schema")
}

func (m *Mixin) Init() error {
	apiServerVersion, err := getKubectlApiServerVersion(m)

	if err != nil {
		return err
	}

	if m.KubernetesClientVersion != apiServerVersion {
		fmt.Fprintf(m.Out, "Kubectl api server version (%s) does not match client version (%s); downloading a compatible client.\n",
			apiServerVersion, m.KubernetesClientVersion)

		err := installKubectlClient(m, apiServerVersion)
		if err != nil {
			return errors.Wrap(err, "unable to install a compatible kubectl client")
		}
	}
	return err
}

func (m *Mixin) getCommandFile(commandFile string, w io.Writer) ([]byte, error) {
	if commandFile == "" {
		reader := bufio.NewReader(m.In)
		return ioutil.ReadAll(reader)
	}
	return ioutil.ReadFile(commandFile)
}

func (m *Mixin) getPayloadData() ([]byte, error) {
	reader := bufio.NewReader(m.In)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		errors.Wrap(err, "could not read payload from STDIN")
	}
	return data, nil
}

func (m *Mixin) ValidatePayload(b []byte) error {
	// Load the step as a go dump
	s := make(map[string]interface{})
	err := yaml.Unmarshal(b, &s)
	if err != nil {
		return errors.Wrap(err, "could not marshal payload as yaml")
	}
	manifestLoader := gojsonschema.NewGoLoader(s)

	// Load the step schema
	schema, err := m.GetSchema()
	if err != nil {
		return err
	}
	schemaLoader := gojsonschema.NewStringLoader(schema)

	validator, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return errors.Wrap(err, "unable to compile the mixin step schema")
	}

	// Validate the manifest against the schema
	result, err := validator.Validate(manifestLoader)
	if err != nil {
		return errors.Wrap(err, "unable to validate the mixin step schema")
	}
	if !result.Valid() {
		errs := make([]string, 0, len(result.Errors()))
		for _, err := range result.Errors() {
			errs = append(errs, err.String())
		}
		return errors.New(strings.Join(errs, "\n\t* "))
	}

	return nil
}

func (m *Mixin) getOutput(resourceType, resourceName, namespace, jsonPath string) ([]byte, error) {
	args := []string{"get", resourceType, resourceName}
	args = append(args, fmt.Sprintf("-o=jsonpath=%s", jsonPath))
	if namespace != "" {
		args = append(args, fmt.Sprintf("--namespace=%s", namespace))
	}
	cmd := m.NewCommand("kubectl", args...)
	cmd.Stderr = m.Err

	prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
	if m.Debug {
		fmt.Fprintln(m.Err, prettyCmd)
	}
	out, err := cmd.Output()

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
	}
	return out, nil
}

func (m *Mixin) handleOutputs(outputs []KubernetesOutput) error {
	//Now get the outputs
	for _, output := range outputs {
		bytes, err := m.getOutput(
			output.ResourceType,
			output.ResourceName,
			output.Namespace,
			output.JSONPath,
		)
		if err != nil {
			return err
		}
		err = m.Context.WriteMixinOutputToFile(output.Name, bytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func installKubectlClient(m *Mixin, version string) error {

	url := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/amd64/kubectl", version)

	// Fetch archive from url
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to construct GET request for fetching kubectl client binary")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to download kubectl client binary via url: %s", url)
	}
	defer res.Body.Close()

	// Create a temp dir
	tmpDir, err := m.FileSystem.TempDir("", "tmp")
	if err != nil {
		return errors.Wrap(err, "unable to create a temporary directory for downloading the kubectl client binary")
	}
	defer os.RemoveAll(tmpDir)

	// Create the local archive
	kubectlBinPath, err := m.FileSystem.Create(filepath.Join(tmpDir, "kubectl"))
	if err != nil {
		return errors.Wrap(err, "unable to create a local file for the kubectl client binary")
	}

	// Copy response body to local archive
	_, err = io.Copy(kubectlBinPath, res.Body)
	if err != nil {
		return errors.Wrap(err, "unable to copy the kubectl client binary to the local archive file")
	}

	// Move the kubectl binary into the appropriate location
	binPath := "/usr/local/bin/kubectl"
	err = m.FileSystem.Rename(fmt.Sprintf("%s", kubectlBinPath), binPath)
	if err != nil {
		return errors.Wrapf(err, "unable to install the kubectl client binary to %q", binPath)
	}
	return nil
}

func getKubectlApiServerVersion(m *Mixin) (string, error) {
	var stderr bytes.Buffer

	cmd := m.NewCommand("kubectl", "api-versions")
	cmd.Stderr = &stderr

	outputBytes, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "unable to determine kubernetes api server version: %s", stderr.String())
	}
	re := regexp.MustCompile(`v[0-9]*\.[0-9]*\.[0-9]*`)
	version := re.FindString(string(outputBytes))

	return version, nil
}
