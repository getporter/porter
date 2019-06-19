package cnabprovider

import (
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pkg/errors"
)

// loadParameters accepts a set of string overrides and combines that with the default parameters to create
// a full set of parameters.
func (d *Duffle) loadParameters(bun *bundle.Bundle, rawOverrides map[string]string) (map[string]interface{}, error) {
	overrides := make(map[string]interface{}, len(rawOverrides))

	for key, rawValue := range rawOverrides {
		paramDef, ok := bun.Parameters.Fields[key]
		if !ok {
			return nil, fmt.Errorf("parameter %s not defined in bundle", key)
		}

		value, err := paramDef.ConvertValue(rawValue)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to convert parameter's %s value %s to the destination parameter type %s", key, rawValue, paramDef.DataType)
		}

		overrides[key] = value
	}

	return bundle.ValuesOrDefaults(overrides, bun)
}
