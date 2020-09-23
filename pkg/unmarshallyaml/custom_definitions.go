// Copyright (c) 2015-2016 Michael Persson
// Copyright (c) 2012–2015 Elasticsearch <http://www.elastic.co>
//
// Originally distributed as part of "beats" repository (https://github.com/elastic/beats).
// Modified specifically for "iodatafmt" package.
//
// Distributed underneath "Apache License, Version 2.0" which is compatible with the LICENSE for this package.
// see https://github.com/go-yaml/yaml/issues/139#issuecomment-220072190

package unmarshallyaml

import (
	"fmt"
)

// CredentialDefinitions allows objects and arrays to be provided as values in custom definitions
// By default yaml serilialiser converts these to map[interface{}]interface which cannot be serialised as json
// see https://github.com/go-yaml/yaml/issues/139

type CustomDefinitions map[string]interface{}

func (cd *CustomDefinitions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[interface{}]interface{}
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	if *cd == nil {
		*cd = make(map[string]interface{}, len(raw))
	}

	for k, v := range raw {
		(*cd)[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}

	return nil
}

func cleanupInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = cleanupMapValue(v)
	}
	return res
}

func cleanupInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	case string, bool, int8, int16, int32, int64, int, uint, uint8, uint16, uint32, uint64:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
