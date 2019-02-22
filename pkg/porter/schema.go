package porter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/context"

	"github.com/deislabs/porter/pkg/mixin"

	"github.com/gobuffalo/packr/v2"
)

func (p *Porter) PrintManifestSchema() error {
	schemaMap, err := p.GetManifestSchema()
	if err != nil {
		return err
	}

	schema, err := json.MarshalIndent(&schemaMap, "", "  ")
	if err != nil {
		return errors.Wrap(err, "could not marshal the composite porter manifest schema")
	}

	fmt.Fprintln(p.Out, string(schema))
	return nil
}

func (p *Porter) GetManifestSchema() (map[string]interface{}, error) {
	t := packr.New("schema", "./schema")

	b, err := t.Find("manifest.json")
	if err != nil {
		return nil, err
	}

	manifestSchema := make(map[string]interface{})
	err = json.Unmarshal(b, &manifestSchema)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal the root porter manifest schema")
	}

	definitionSchema, ok := manifestSchema["definitions"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid definitions type, expected map[string]interface{} but got %T", manifestSchema["definitions"])
	}

	propertiesSchema, ok := manifestSchema["properties"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties type, expected map[string]interface{} but got %T", manifestSchema["properties"])
	}

	mixinSchema, ok := propertiesSchema["mixins"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties.mixins type, expected map[string]interface{} but got %T", propertiesSchema["mixins"])
	}

	itemsSchema, ok := mixinSchema["items"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties.mixins.items type, expected map[string]interface{} but got %T", mixinSchema["items"])
	}

	enumSchema, ok := itemsSchema["enum"].([]interface{})
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties.mixins.items.enum type, expected []interface{} but got %T", itemsSchema["enum"])
	}

	installSchema, ok := propertiesSchema["install"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties.install type, expected map[string]interface{} but got %T", propertiesSchema["install"])
	}

	anyOfSchema, ok := installSchema["anyOf"].([]interface{})
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties.install.anyOf type, expected []interface{} but got %T", installSchema["anyOf"])
	}

	mixins, err := p.GetMixins()
	if err != nil {
		return nil, err
	}

	for _, mixin := range mixins {
		mixinSchema, err := p.getMixinSchema(mixin)
		if err != nil {
			// if a mixin can't report its schema, don't include it and keep going
			if p.Debug {
				fmt.Fprintln(p.Err, errors.Wrapf(err, "could not query mixin %s for its schema", mixin.Name))
			}
			continue
		}

		for action, actionSchema := range mixinSchema {
			if !config.IsSupportedAction(action) {
				continue
			}

			ref := fmt.Sprintf("%s.%s", mixin.Name, action)
			definitionSchema[ref] = actionSchema
			enumSchema = append(enumSchema, mixin.Name)
			anyOfSchema = append(anyOfSchema, map[string]interface{}{"$ref": "#/definitions/" + ref})
		}
	}
	// Save the updated arrays into the json schema document
	itemsSchema["enum"] = enumSchema
	installSchema["anyOf"] = anyOfSchema

	return manifestSchema, nil
}

func (p *Porter) getMixinSchema(m mixin.Metadata) (map[string]interface{}, error) {
	r := mixin.NewRunner(m.Name, m.Dir, false)
	r.Command = "schema"

	// Copy the existing context and tweak to pipe the output differently
	mixinSchema := &bytes.Buffer{}
	var mixinContext context.Context
	mixinContext = *p.Context
	mixinContext.Out = mixinSchema
	if !p.Debug {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	err := r.Run()
	if err != nil {
		return nil, err
	}

	schemaMap := make(map[string]interface{})
	err = json.Unmarshal(mixinSchema.Bytes(), &schemaMap)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal mixin schema for %s, %q", m.Name, mixinSchema.String())
	}

	return schemaMap, nil
}
