//go:generate packr2

package kubernetes

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
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
