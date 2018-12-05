package azure

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type InstallArguments struct {
	Type          string                 `yaml:"type"`
	Template      string                 `yaml:"template"`
	Name          string                 `yaml:"name"`
	ResourceGroup string                 `yaml:"resourceGroup"`
	Parameters    map[string]interface{} `yaml:"parameters"`
}

func (m *Mixin) Install() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}
	var args InstallArguments
	err = yaml.Unmarshal(payload, &args)
	if err != nil {
		return err
	}

	// Get the arm deployer

	deployer, err := m.getARMDeployer()
	// Get the Template based on the arguments (type)
	t, err := deployer.FindTemplate(args.Type, args.Template)
	if err != nil {
		return err
	}

	// call Deployer.Deploy(...)
	outputs, err := deployer.Deploy(
		args.Name,
		args.ResourceGroup,
		args.Parameters["location"].(string),
		t,
		args.Parameters, //arm params
		nil,             //Tags not supported right now
	)
	if err != nil {
		return err
	}

	f, err := m.Context.NewOutput()
	if err != nil {
		return err
	}
	defer f.Close()
	for k, v := range outputs {
		val, ok := v.(string)
		if !ok {
			val = fmt.Sprintf("%v", v)
		}
		os.Setenv(k, val)
	}
	return nil
}
