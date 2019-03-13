//go:generate packr2

package kubernetes

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/deislabs/porter/pkg/context"
	"github.com/ghodss/yaml"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

const defaultManifestPath = "/cnab/app/manifests/kubernetes"

type Mixin struct {
	*context.Context

	schemas *packr.Box
}

func New() *Mixin {
	return &Mixin{
		Context: context.New(),
		schemas: NewSchemaBox(),
	}
}

func NewSchemaBox() *packr.Box {
	return packr.New("github.com/deislabs/porter/pkg/kubernetes/schema", "./schema")
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
	return data, errors.Wrap(err, "could not read payload from STDIN")
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

// If no manifest is specified, update the empty slice to include the default path
func (m *Mixin) resolveManifests(manifests []string) []string {
	if len(manifests) == 0 {
		return append(manifests, defaultManifestPath)
	}
	return manifests
}

func (m *Mixin) getOutput(resourceType, resourceName, namespace, jsonPath string) (string, error) {
	args := []string{"get", resourceType, resourceName}
	args = append(args, fmt.Sprintf("-o=jsonpath='%s'", jsonPath))
	if namespace != "" {
		args = append(args, fmt.Sprintf("--namespace=%s", namespace))
	}
	cmd := m.NewCommand("kubectl", args...)
	cmd.Stderr = m.Err
	out, err := cmd.Output()
	if err != nil {
		prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
		return "", errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
	}
	return string(out), nil
}

func (m *Mixin) handleOutputs(outputs []KubernetesOutput) error {
	//Now get the outputs
	var lines []string
	for _, output := range outputs {
		val, err := m.getOutput(
			output.ResourceType,
			output.ResourceName,
			output.Namespace,
			output.JSONPath,
		)
		if err != nil {
			return err
		}
		l := fmt.Sprintf("%s=%s", output.Name, val)
		lines = append(lines, l)
	}
	m.Context.WriteOutput(lines)
	return nil
}
