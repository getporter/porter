package porter

import (
	"encoding/json"
	"fmt"

	"github.com/deislabs/porter/pkg/config"

	"github.com/pkg/errors"
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
	b, err := p.schemas.Find("manifest.json")
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
		mixinSchema, err := p.GetMixinSchema(mixin)
		if err != nil {
			// if a mixin can't report its schema, don't include it and keep going
			if p.Debug {
				fmt.Fprintln(p.Err, errors.Wrapf(err, "could not query mixin %s for its schema", mixin.Name))
			}
			continue
		}

		mixinSchemaMap := make(map[string]interface{})
		err = json.Unmarshal([]byte(mixinSchema), &mixinSchemaMap)
		if err != nil {
			return nil, errors.Wrapf(err, "could not unmarshal mixin schema for %s, %q", mixin.Name, mixinSchema)
		}

		for _, action := range config.GetSupportActions() {
			actionSchema := mixinSchemaMap[string(action)]

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
