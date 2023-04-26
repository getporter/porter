package encoding

import (
	"encoding/json"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArrayMap_Merge(t *testing.T) {
	m := &ArrayMap[TestMapEntry, TestArrayEntry]{}
	m.Set("first", TestMapEntry{Value: "base first"})
	m.Set("second", TestMapEntry{Value: "base second"})
	m.Set("third", TestMapEntry{Value: "base third"})

	result := m.Merge(&ArrayMap[TestMapEntry, TestArrayEntry]{})
	require.Equal(t, 3, result.Len())
	_, ok := result.Get("fourth")
	assert.False(t, ok, "fourth should not be present in the base set")

	wantFourth := TestMapEntry{Value: "new fourth"}
	fourthMap := &ArrayMap[TestMapEntry, TestArrayEntry]{items: map[string]TestMapEntry{"fourth": wantFourth}}
	result = m.Merge(fourthMap)
	require.Equal(t, 4, result.Len())
	gotFourth, ok := result.Get("fourth")
	require.True(t, ok, "fourth should be present in the merged set")
	assert.Equal(t, wantFourth, gotFourth, "incorrect merged value for fourth")

	wantSecond := TestMapEntry{Value: "override second"}
	secondMap := &ArrayMap[TestMapEntry, TestArrayEntry]{items: map[string]TestMapEntry{"second": wantSecond}}
	result = m.Merge(secondMap)
	require.Equal(t, 3, result.Len())
	gotSecond, ok := result.Get("second")
	require.True(t, ok, "second should be present in the merged set")
	assert.Equal(t, wantSecond, gotSecond, "incorrect merged value for second")
}

func TestArrayMap_Unmarshal(t *testing.T) {
	// TODO: add testcase for json
	data, err := os.ReadFile("testdata/array.yaml")
	require.NoError(t, err, "ReadFile failed")

	var m ArrayMap[TestMapEntry, TestArrayEntry]
	err = yaml.Unmarshal(data, &m)
	require.NoError(t, err, "Unmarshal failed")

	require.Equal(t, 2, m.Len(), "unexpected number of items defined")

	gotA, ok := m.Get("aname")
	require.True(t, ok, "aname was not defined")
	wantA := TestMapEntry{Value: "stuff"}
	assert.Equal(t, wantA, gotA, "unexpected aname defined")

	gotB, ok := m.Get("bname")
	require.True(t, ok, "password was not defined")
	wantB := TestMapEntry{Value: "things"}
	assert.Equal(t, wantB, gotB, "unexpected bname defined")
}

func TestArrayMap_Marshal(t *testing.T) {
	// TODO: add testcase for json

	m := &ArrayMap[TestMapEntry, TestArrayEntry]{}
	m.Set("bname", TestMapEntry{Value: "things"})
	m.Set("aname", TestMapEntry{Value: "stuff"})

	wantData, err := os.ReadFile("testdata/array.yaml")
	require.NoError(t, err, "ReadFile failed")

	gotData, err := yaml.Marshal(m)
	require.NoError(t, err, "Marshal failed")
	assert.Equal(t, string(wantData), string(gotData))
}

func TestArrayMap_Unmarshal_DuplicateKeys(t *testing.T) {
	data, err := os.ReadFile("testdata/array-with-duplicates.yaml")
	require.NoError(t, err, "ReadFile failed")

	var l ArrayMap[TestMapEntry, TestArrayEntry]
	err = yaml.Unmarshal(data, &l)
	require.ErrorContains(t, err, "cannot unmarshal source map: duplicate key found 'aname'")
}

// check that when we round trip a null ArrayMap, it stays null and isn't initialized to an _empty_ ArrayMap
// This impacts how it is marshaled later to yaml or json, because we often have fields tagged with omitempty
// and so it must be null to not be written out.
func TestArrayMap_RoundTrip_Empty(t *testing.T) {
	wantData, err := os.ReadFile("testdata/array-empty.json")
	require.NoError(t, err, "ReadFile failed")

	var s struct {
		Items *ArrayMap[TestMapEntry, TestArrayEntry] `json:"items"`
	}
	s.Items = &ArrayMap[TestMapEntry, TestArrayEntry]{}

	gotData, err := json.Marshal(s)
	require.NoError(t, err, "Marshal failed")
	require.Equal(t, string(wantData), string(gotData), "empty ArrayMap should not marshal as empty, but nil so that it works with omitempty")

	err = json.Unmarshal(gotData, &s)
	require.NoError(t, err, "Unmarshal failed")
	require.Nil(t, s.Items, "null ArrayMap should unmarshal as nil")
}

type TestMapEntry struct {
	Value string `json:"value" yaml:"value"`
}

func (t TestMapEntry) ToArrayEntry(key string) ArrayElement {
	return TestArrayEntry{Name: key, Value: t.Value}
}

type TestArrayEntry struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

func (t TestArrayEntry) ToMapEntry() MapElement {
	return TestMapEntry{Value: t.Value}
}

func (t TestArrayEntry) GetKey() string {
	return t.Name
}
