package porter

import (
	"encoding/json"
	"fmt"

	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

type jsonSchema = map[string]interface{}
type jsonObject = map[string]interface{}

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

func (p *Porter) GetManifestSchema() (jsonSchema, error) {
	b, err := p.Templates.GetSchemaTemplate()
	if err != nil {
		return nil, err
	}

	manifestSchema := make(jsonSchema)
	err = json.Unmarshal(b, &manifestSchema)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal the root porter manifest schema")
	}

	definitionSchema, ok := manifestSchema["definitions"].(jsonSchema)
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid definitions type, expected map[string]interface{} but got %T", manifestSchema["definitions"])
	}

	propertiesSchema, ok := manifestSchema["properties"].(jsonSchema)
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties type, expected map[string]interface{} but got %T", manifestSchema["properties"])
	}

	mixinSchema, ok := propertiesSchema["mixins"].(jsonSchema)
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties.mixins type, expected map[string]interface{} but got %T", propertiesSchema["mixins"])
	}

	mixinItemSchema, ok := mixinSchema["items"].(jsonSchema)
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties.mixins.items type, expected map[string]interface{} but got %T", mixinSchema["items"])
	}

	mixinEnumSchema, ok := mixinItemSchema["enum"].([]interface{})
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties.mixins.items.enum type, expected []interface{} but got %T", mixinItemSchema["enum"])
	}

	supportedActions := config.GetSupportActions()
	actionSchemas := make(map[string]jsonSchema, len(supportedActions))
	for _, action := range supportedActions {
		actionSchema, ok := propertiesSchema[string(action)].(jsonSchema)
		if !ok {
			return nil, errors.Errorf("root porter manifest schema has invalid properties.%s type, expected map[string]interface{} but got %T", action, propertiesSchema[string(action)])
		}
		actionSchemas[string(action)] = actionSchema
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

		mixinEnumSchema = append(mixinEnumSchema, mixin.Name)

		mixinSchemaMap := make(jsonSchema)
		err = json.Unmarshal([]byte(mixinSchema), &mixinSchemaMap)
		if err != nil {
			return nil, errors.Wrapf(err, "could not unmarshal mixin schema for %s, %q", mixin.Name, mixinSchema)
		}

		for _, action := range config.GetSupportActions() {
			mixinActionSchema := mixinSchemaMap[string(action)]

			ref := fmt.Sprintf("%s.%s", mixin.Name, action)
			definitionSchema[ref] = mixinActionSchema

			actionItemSchema, ok := actionSchemas[string(action)]["items"].(jsonSchema)
			if !ok {
				return nil, errors.Errorf("root porter manifest schema has invalid properties.%s.items type, expected map[string]interface{} but got %T", action, actionSchemas[string(action)]["items"])
			}
			actionAnyOfSchema, ok := actionItemSchema["anyOf"].([]interface{})
			if !ok {
				return nil, errors.Errorf("root porter manifest schema has invalid properties.%s.items.anyOf type, expected []interface{} but got %T", action, actionItemSchema["anyOf"])
			}
			actionAnyOfSchema = append(actionAnyOfSchema, jsonObject{"$ref": "#/definitions/" + ref})
			actionItemSchema["anyOf"] = actionAnyOfSchema
		}
	}

	// Save the updated arrays into the json schema document
	mixinItemSchema["enum"] = mixinEnumSchema

	return manifestSchema, nil
}
