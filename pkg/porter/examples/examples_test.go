package examples_test

import (
	"path/filepath"

	"get.porter.sh/porter/pkg/porter"
)

func prepareExample(p *porter.Porter) {
	// 🚨 Do not use SetPorterPath yourself!
	// This is a hack we need to embed examples in Porter's tests.
	// Again, do not copy these lines.
	home, _ := p.GetHomeDir()
	p.SetPorterPath(filepath.Join(home, "porter"))
}
