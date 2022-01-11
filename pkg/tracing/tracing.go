package tracing

import (
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ConvertAttributesToFields(attrs []attribute.KeyValue) []zap.Field {
	fields := make([]zap.Field, len(attrs))
	for i, attr := range attrs {
		fields[i] = ConvertAttributeToField(attr)
	}
	return fields
}

func ConvertAttributeToField(attr attribute.KeyValue) zapcore.Field {
	key := string(attr.Key)

	switch attr.Value.Type() {

	case attribute.BOOL:
		return zap.Bool(key, attr.Value.AsBool())
	case attribute.BOOLSLICE:
		return zap.Bools(key, attr.Value.AsBoolSlice())
	case attribute.FLOAT64:
		return zap.Float64(key, attr.Value.AsFloat64())
	case attribute.FLOAT64SLICE:
		return zap.Float64s(key, attr.Value.AsFloat64Slice())
	case attribute.INT64:
		return zap.Int64(key, attr.Value.AsInt64())
	case attribute.INT64SLICE:
		return zap.Int64s(key, attr.Value.AsInt64Slice())
	case attribute.STRING:
		return zap.String(key, attr.Value.AsString())
	case attribute.STRINGSLICE:
		return zap.Strings(key, attr.Value.AsStringSlice())
	default:
		// Give up and do our best
		b, _ := attr.Value.MarshalJSON()
		return zap.String(key, string(b))
	}
}
