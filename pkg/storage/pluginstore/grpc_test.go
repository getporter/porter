package pluginstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestConvertFloatToInt(t *testing.T) {
	src := map[string]interface{}{
		"a": map[string]interface{}{
			"b": float64(1),
			"c": []interface{}{
				float64(1),
				float64(-1),
			},
		},
		"d": []interface{}{
			map[string]interface{}{"e": "cat"},
		},
	}

	dest := ConvertFloatToInt(src)

	wantDest := map[string]interface{}{
		"a": map[string]interface{}{
			"b": int64(1),
			"c": []interface{}{int64(1), int64(-1)}},
		"d": []interface{}{map[string]interface{}{
			"e": "cat"},
		},
	}
	assert.Equal(t, wantDest, dest)
}

func TestConvertBsonM(t *testing.T) {
	// Check that AsMap fixes float->int
	src := map[string]interface{}{
		"a": map[string]interface{}{
			"b": float64(1),
			"c": []interface{}{
				float64(1),
				float64(-1),
			},
		},
	}

	tmp := NewStruct(src)
	dest := AsMap(tmp)

	wantDest := bson.M{
		"a": map[string]interface{}{ // right now we only convert the top level to the expected bson type. Mongo doesn't care if farther down we use primitives
			"b": int64(1),
			"c": []interface{}{
				int64(1),
				int64(-1),
			},
		},
	}
	require.Equal(t, wantDest, dest)
}

func TestConvertBsonD(t *testing.T) {
	src := bson.D{
		{Key: "a", Value: "1"},
		{Key: "b", Value: bson.D{
			{Key: "c", Value: 1},
		}},
	}

	tmp := FromOrderedMap(src)
	dest := AsOrderedMap(tmp, ConvertSliceToBsonD)

	wantDest := bson.D{
		{Key: "a", Value: "1"},
		{Key: "b", Value: bson.D{{Key: "c", Value: int64(1)}}},
	}
	require.Equal(t, wantDest, dest)
}
