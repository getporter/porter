package mixin

import (
	"io/ioutil"

	"github.com/deislabs/porter/pkg/context"
)

// TestMixinProvider helps us test Porter.Mixins in our unit tests without actually hitting any real mixins on the file system.
type TestMixinProvider struct {
	RunAssertions []func(mixinCxt *context.Context, mixinName string, commandOpts CommandOptions)
}

func (p *TestMixinProvider) List() ([]Metadata, error) {
	mixins := []Metadata{
		{Name: "exec"},
	}
	return mixins, nil
}

func (p *TestMixinProvider) GetSchema(m Metadata) (string, error) {
	b, err := ioutil.ReadFile("../exec/schema/exec.json")
	return string(b), err
}

func (p *TestMixinProvider) GetVersion(m Metadata) (string, error) {
	return "exec mixin v1.0 (abc123)", nil
}

func (p *TestMixinProvider) GetVersionMetadata(m Metadata) (*VersionInfo, error) {
	return &VersionInfo{Version: "v1.0", Commit: "abc123", Author: "Deis Labs"}, nil
}

func (p *TestMixinProvider) Install(o InstallOptions) (*Metadata, error) {
	return &Metadata{Name: "exec", Dir: "~/.porter/mixins/exec"}, nil
}

func (p *TestMixinProvider) Uninstall(o UninstallOptions) (*Metadata, error) {
	return &Metadata{Name: "exec"}, nil
}

func (p *TestMixinProvider) Run(mixinCxt *context.Context, mixinName string, commandOpts CommandOptions) error {
	for _, assert := range p.RunAssertions {
		assert(mixinCxt, mixinName, commandOpts)
	}
	return nil
}
