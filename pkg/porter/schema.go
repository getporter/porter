package porter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/PaesslerAG/jsonpath"
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
	replacementSchema, err := p.GetReplacementSchema()
	if err != nil && p.Debug {
		fmt.Fprintln(p.Err, errors.Wrap(err, "ignoring replacement schema"))
	}
	if replacementSchema != nil {
		return replacementSchema, nil
	}

	b, err := p.Templates.GetSchema()
	if err != nil {
		return nil, err
	}

	manifestSchema := make(jsonSchema)
	err = json.Unmarshal(b, &manifestSchema)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal the root porter manifest schema")
	}

	combinedSchema, err := p.injectMixinSchemas(manifestSchema)
	if err != nil {
		if p.Debug {
			fmt.Fprintln(p.Err, err)
		}
		// Fallback to the porter schema, without any mixins
		return manifestSchema, nil
	}

	return combinedSchema, nil
}

func (p *Porter) injectMixinSchemas(manifestSchema jsonSchema) (jsonSchema, error) {
	propertiesSchema, ok := manifestSchema["properties"].(jsonSchema)
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid properties type, expected map[string]interface{} but got %T", manifestSchema["properties"])
	}

	additionalPropertiesSchema, ok := manifestSchema["additionalProperties"].(jsonSchema)
	if !ok {
		return nil, errors.Errorf("root porter manifest schema has invalid additionalProperties type, expected map[string]interface{} but got %T", manifestSchema["additionalProperties"])
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

	coreActions := []string{"install", "upgrade", "uninstall"} // custom actions are defined in json schema as additionalProperties
	actionSchemas := make(map[string]jsonSchema, len(coreActions)+1)
	for _, action := range coreActions {
		actionSchema, ok := propertiesSchema[action].(jsonSchema)
		if !ok {
			return nil, errors.Errorf("root porter manifest schema has invalid properties.%s type, expected map[string]interface{} but got %T", action, propertiesSchema[string(action)])
		}
		actionSchemas[action] = actionSchema
	}

	mixins, err := p.Mixins.List()
	if err != nil {
		return nil, err
	}

	// If there is an error with any mixin, print a warning and skip the mixin, do not return an error
	for _, mixin := range mixins {
		mixinSchema, err := p.Mixins.GetSchema(mixin)
		if err != nil {
			// if a mixin can't report its schema, don't include it and keep going
			if p.Debug {
				fmt.Fprintln(p.Err, errors.Wrapf(err, "could not query mixin %s for its schema", mixin))
			}
			continue
		}

		// Update relative refs with the new location and reload
		mixinSchema = strings.Replace(mixinSchema, "#/", fmt.Sprintf("#/mixin.%s/", mixin), -1)

		mixinSchemaMap := make(jsonSchema)
		err = json.Unmarshal([]byte(mixinSchema), &mixinSchemaMap)
		if err != nil && p.Debug {
			fmt.Fprintln(p.Err, errors.Wrapf(err, "could not unmarshal mixin schema for %s, %q", mixin, mixinSchema))
			continue
		}

		mixinEnumSchema = append(mixinEnumSchema, mixin)

		// embed the entire mixin schema in the root
		manifestSchema["mixin."+mixin] = mixinSchemaMap

		for _, action := range coreActions {
			actionItemSchema, ok := actionSchemas[action]["items"].(jsonSchema)
			if err != nil && p.Debug {
				fmt.Fprintln(p.Err, errors.Errorf("root porter manifest schema has invalid properties.%s.items type, expected map[string]interface{} but got %T", action, actionSchemas[string(action)]["items"]))
				continue
			}

			actionAnyOfSchema, ok := actionItemSchema["anyOf"].([]interface{})
			if !ok {
				if err != nil && p.Debug {
					fmt.Fprintln(p.Err, errors.Errorf("root porter manifest schema has invalid properties.%s.items.anyOf type, expected []interface{} but got %T", action, actionItemSchema["anyOf"]))
					continue
				}
			}

			actionRef := fmt.Sprintf("#/mixin.%s/definitions/%sStep", mixin, action)
			actionAnyOfSchema = append(actionAnyOfSchema, jsonObject{"$ref": actionRef})
			actionItemSchema["anyOf"] = actionAnyOfSchema
		}

		// Some mixins don't support custom actions, if the mixin has invokeStep defined,
		// then use it in our additionalProperties list of acceptable root level elements.
		_, err = jsonpath.Get("$.definitions.invokeStep", mixinSchemaMap)
		if err == nil {
			actionItemSchema, ok := additionalPropertiesSchema["items"].(jsonSchema)
			if err != nil && p.Debug {
				fmt.Fprintln(p.Err, errors.Errorf("root porter manifest schema has invalid additionalProperties.items type, expected map[string]interface{} but got %T", additionalPropertiesSchema["items"]))
				continue
			}

			actionAnyOfSchema, ok := actionItemSchema["anyOf"].([]interface{})
			if !ok && p.Debug {
				fmt.Fprintln(p.Err, errors.Errorf("root porter manifest schema has invalid additionalProperties.items.anyOf type, expected []interface{} but got %T", actionItemSchema["anyOf"]))
				continue
			}

			actionRef := fmt.Sprintf("#/mixin.%s/definitions/invokeStep", mixin)
			actionAnyOfSchema = append(actionAnyOfSchema, jsonObject{"$ref": actionRef})
			actionItemSchema["anyOf"] = actionAnyOfSchema
		}
	}

	// Save the updated arrays into the json schema document
	mixinItemSchema["enum"] = mixinEnumSchema

	return manifestSchema, err
}

func (p *Porter) GetReplacementSchema() (jsonSchema, error) {
	home, err := p.GetHomeDir()
	if err != nil {
		return nil, err
	}

	replacementSchemaPath := filepath.Join(home, "porter.json")
	if exists, _ := p.FileSystem.Exists(replacementSchemaPath); !exists {
		return nil, nil
	}

	b, err := p.FileSystem.ReadFile(replacementSchemaPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read replacement schema at %s", replacementSchemaPath)
	}

	replacementSchema := make(jsonSchema)
	err = json.Unmarshal(b, &replacementSchema)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal replacement schema in %s", replacementSchemaPath)
	}

	return replacementSchema, nil
}
