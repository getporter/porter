package tracing

import (
	"encoding/json"

	"go.opentelemetry.io/otel/attribute"
)

// ObjectAttribute writes the specified value as a json encoded value.
func ObjectAttribute(key string, value interface{}) attribute.KeyValue {
	data, err := json.Marshal(value)
	if err != nil {
		data = []byte("unable to marshal object to an otel attribute")
	}

	return attribute.Key(key).String(string(data))
}
