// Copyright (c) 2015-2016 Michael Persson
// Copyright (c) 2012–2015 Elasticsearch <http://www.elastic.co>
//
// Originally distributed as part of "beats" repository (https://github.com/elastic/beats).
// Modified specifically for "iodatafmt" package.
// Modified to make UnmarshalYAML reusable by other types, turn null keys to "null" and be compatible with gopkg.in/yaml.v3
//
// Distributed underneath "Apache License, Version 2.0" which is compatible with the LICENSE for this package.
// see https://github.com/go-yaml/yaml/issues/139#issuecomment-220072190

package yaml

import (
	"fmt"
)

// UnmarshalMap allows unmarshaling into types that are safe to then be marshaled to json.
// By default yaml serializer converts these to map[interface{}]interface which cannot be serialised as json
// see https://github.com/go-yaml/yaml/issues/139
// This is still required even with v3 because if any key isn't parsed as a string, e.g. you have a key named 1 or 2, then the bug is still triggered.
func UnmarshalMap(unmarshal func(interface{}) error) (map[string]interface{}, error) {
	var raw map[string]interface{}
	err := unmarshal(&raw)
	if err != nil {
		return nil, err
	}

	for k, v := range raw {
		raw[stringKey(k)] = cleanupMapValue(v)
	}

	return raw, nil
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
		res[stringKey(k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupStringMap(in map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[stringKey(k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	case map[string]interface{}:
		return cleanupStringMap(v)
	default:
		return v
	}
}

func stringKey(k interface{}) string {
	switch k {
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", k)
	}
}
