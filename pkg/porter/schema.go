package porter

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/tracing"
	"github.com/PaesslerAG/jsonpath"
)

type jsonSchema = map[string]interface{}
type jsonObject = map[string]interface{}

func (p *Porter) PrintManifestSchema(ctx context.Context) error {
	schemaMap, err := p.GetManifestSchema(ctx)
	if err != nil {
		return err
	}

	schema, err := json.MarshalIndent(&schemaMap, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal the composite porter manifest schema: %w", err)
	}

	fmt.Fprintln(p.Out, string(schema))
	return nil
}

func (p *Porter) GetManifestSchema(ctx context.Context) (jsonSchema, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	replacementSchema, err := p.GetReplacementSchema()
	if err != nil {
		span.Debugf("ignoring replacement schema: %w", err)
	}
	if replacementSchema != nil {
		return replacementSchema, nil
	}

	b, err := p.Templates.GetSchema()
	if err != nil {
		return nil, span.Error(err)
	}

	manifestSchema := make(jsonSchema)
	err = json.Unmarshal(b, &manifestSchema)
	if err != nil {
		return nil, span.Error(fmt.Errorf("could not unmarshal the root porter manifest schema: %w", err))
	}

	combinedSchema, err := p.injectMixinSchemas(ctx, manifestSchema)
	if err != nil {
		span.Warn(err.Error())
		// Fallback to the porter schema, without any mixins
		return manifestSchema, nil
	}

	return combinedSchema, nil
}

func (p *Porter) injectMixinSchemas(ctx context.Context, manifestSchema jsonSchema) (jsonSchema, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	propertiesSchema, ok := manifestSchema["properties"].(jsonSchema)
	if !ok {
		return nil, span.Error(fmt.Errorf("root porter manifest schema has invalid properties type, expected map[string]interface{} but got %T", manifestSchema["properties"]))
	}

	additionalPropertiesSchema, ok := manifestSchema["additionalProperties"].(jsonSchema)
	if !ok {
		return nil, span.Error(fmt.Errorf("root porter manifest schema has invalid additionalProperties type, expected map[string]interface{} but got %T", manifestSchema["additionalProperties"]))
	}

	mixinSchema, ok := propertiesSchema["mixins"].(jsonSchema)
	if !ok {
		return nil, span.Error(fmt.Errorf("root porter manifest schema has invalid properties.mixins type, expected map[string]interface{} but got %T", propertiesSchema["mixins"]))
	}

	mixinItemSchema, ok := mixinSchema["items"].(jsonSchema)
	if !ok {
		return nil, span.Error(fmt.Errorf("root porter manifest schema has invalid properties.mixins.items type, expected map[string]interface{} but got %T", mixinSchema["items"]))
	}

	// the set of acceptable ways to declare a mixin
	// e.g.
	// mixins:
	// - exec
	// - helm3:
	//     clientVersion: 1.2.3
	mixinDeclSchema, ok := mixinItemSchema["oneOf"].([]interface{})
	if !ok {
		return nil, span.Error(fmt.Errorf("root porter manifest schema has invalid properties.mixins.items.oneOf type, expected []interface{} but got %T", mixinItemSchema["oneOf"]))
	}

	// The first item is an enum of all the mixin names
	if len(mixinDeclSchema) > 1 {
		return nil, span.Errorf("root porter manifest schema has invalid properties.mixins.items.oneOf, expected a string type to list the names of all the mixins")
	}
	mixinNameDecl, ok := mixinDeclSchema[0].(jsonSchema)
	if !ok {
		return nil, span.Error(fmt.Errorf("root porter manifest schema has invalid properties.mixins.items.oneOf[0] type, expected []map[string]interface{} but got %T", mixinNameDecl))
	}
	mixinNameEnum, ok := mixinNameDecl["enum"].([]interface{})
	if !ok {
		return nil, span.Error(fmt.Errorf("root porter manifest schema has invalid properties.mixins.items.oneOf[0].enum type, expected []interface{} but got %T", mixinNameDecl["enum"]))
	}

	coreActions := []string{"install", "upgrade", "uninstall"} // custom actions are defined in json schema as additionalProperties
	actionSchemas := make(map[string]jsonSchema, len(coreActions)+1)
	for _, action := range coreActions {
		actionSchema, ok := propertiesSchema[action].(jsonSchema)
		if !ok {
			return nil, span.Error(fmt.Errorf("root porter manifest schema has invalid properties.%s type, expected map[string]interface{} but got %T", action, propertiesSchema[string(action)]))
		}
		actionSchemas[action] = actionSchema
	}

	mixins, err := p.Mixins.List()
	if err != nil {
		return nil, span.Error(err)
	}

	// If there is an error with any mixin, print a warning and skip the mixin, do not return an error
	for _, mixin := range mixins {
		mixinSchema, err := p.Mixins.GetSchema(ctx, mixin)
		if err != nil {
			// if a mixin can't report its schema, don't include it and keep going
			span.Debugf("could not query mixin %s for its schema: %w", mixin, err)
			continue
		}

		// Update relative refs with the new location and reload
		mixinSchema = strings.Replace(mixinSchema, "#/", fmt.Sprintf("#/mixin.%s/", mixin), -1)

		mixinSchemaMap := make(jsonSchema)
		err = json.Unmarshal([]byte(mixinSchema), &mixinSchemaMap)
		if err != nil {
			span.Debugf("could not unmarshal mixin schema for %s, %q: %w", mixin, mixinSchema, err)
			continue
		}

		// Support declaring the mixin just by name
		mixinNameEnum = append(mixinNameEnum, mixin)

		// Support configuring the mixin, if available
		// We know it's supported if it has a config definition included in its schema
		if _, err := jsonpath.Get("$.definitions.config", mixinSchemaMap); err == nil {
			mixinConfigRef := fmt.Sprintf("#/mixin.%s/definitions/config", mixin)
			mixinDeclSchema = append(mixinDeclSchema, jsonObject{"$ref": mixinConfigRef})
		}

		// embed the entire mixin schema in the root
		manifestSchema["mixin."+mixin] = mixinSchemaMap

		for _, action := range coreActions {
			actionItemSchema, ok := actionSchemas[action]["items"].(jsonSchema)
			if !ok {
				span.Debugf("root porter manifest schema has invalid properties.%s.items type, expected map[string]interface{} but got %T", action, actionSchemas[string(action)]["items"])
				continue
			}

			actionAnyOfSchema, ok := actionItemSchema["anyOf"].([]interface{})
			if !ok {
				span.Debugf("root porter manifest schema has invalid properties.%s.items.anyOf type, expected []interface{} but got %T", action, actionItemSchema["anyOf"])
				continue
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
			if !ok {
				span.Debugf("root porter manifest schema has invalid additionalProperties.items type, expected map[string]interface{} but got %T", additionalPropertiesSchema["items"])
				continue
			}

			actionAnyOfSchema, ok := actionItemSchema["anyOf"].([]interface{})
			if !ok {
				span.Debugf("root porter manifest schema has invalid additionalProperties.items.anyOf type, expected []interface{} but got %T", actionItemSchema["anyOf"])
				continue
			}

			actionRef := fmt.Sprintf("#/mixin.%s/definitions/invokeStep", mixin)
			actionAnyOfSchema = append(actionAnyOfSchema, jsonObject{"$ref": actionRef})
			actionItemSchema["anyOf"] = actionAnyOfSchema
		}
	}

	// Save the updated arrays into the json schema document
	mixinNameDecl["enum"] = mixinNameEnum
	mixinItemSchema["oneOf"] = mixinDeclSchema

	return manifestSchema, span.Error(err)
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
		return nil, fmt.Errorf("could not read replacement schema at %s: %w", replacementSchemaPath, err)
	}

	replacementSchema := make(jsonSchema)
	err = json.Unmarshal(b, &replacementSchema)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal replacement schema in %s: %w", replacementSchemaPath, err)
	}

	return replacementSchema, nil
}
