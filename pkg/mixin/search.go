package mixin

import (
	"encoding/json"
	"strings"

	"get.porter.sh/porter/pkg/printer"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
)

// SearchOptions are options for searching remote mixins
type SearchOptions struct {
	Name string
	printer.PrintOptions
}

// Searcher contains a packr.Box containing a searchable list of mixins
type Searcher struct {
	Box *packr.Box
}

// NewSearcher returns a Searcher with the provided packr.Box
func NewSearcher(box *packr.Box) Searcher {
	return Searcher{
		Box: box,
	}
}

// Validate validates the arguments provided to a search command
func (o *SearchOptions) Validate(args []string) error {
	err := o.validateMixinName(args)
	if err != nil {
		return err
	}

	return o.ParseFormat()
}

// validateMixinName validates either no mixin name is provided or only one is
func (o *SearchOptions) validateMixinName(args []string) error {
	switch len(args) {
	case 0:
		return nil
	case 1:
		o.Name = strings.ToLower(args[0])
		return nil
	default:
		return errors.Errorf("only one positional argument may be specified, the mixin name, but multiple were received: %s", args)
	}
}

// Search searches for mixins matching the optional provided name,
// returning the full list if none is provided
func (m *Searcher) Search(opts SearchOptions) ([]RemoteMixinInfo, error) {
	data, err := m.Box.Find("index.json")
	if err != nil {
		return nil, errors.Wrap(err, "error loading remote mixin list")
	}

	var rmis []PackageListing
	err = json.Unmarshal(data, &rmis)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse remote mixin list")
	}

	results := make([]PackageListing, len(rmis))
	copy(results, rmis)

	if opts.Name != "" {
		results = []PackageListing{}
		for _, rmi := range rmis {
			if strings.Contains(rmi.Name, opts.Name) {
				results = append(results, rmi)
			}
		}
	}

	return results, nil
}
