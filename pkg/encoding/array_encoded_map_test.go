package encoding

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/carolynvs/aferox"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeArrayEncodedMap(t *testing.T) {
	m := MakeArrayEncodedMap[TestMapEntry, TestArrayEntry](5)
	require.NotNil(t, m.items, "MakeArrayEncodedMap should initialize the backing items")
	// go doesn't let us read back out the capacity
}

func TestNewArrayEncodedMap(t *testing.T) {
	m := NewArrayEncodedMap[TestMapEntry, TestArrayEntry]()
	require.NotNil(t, m.items, "NewArrayEncodedMap should initialize the backing items")
	require.Empty(t, m.items, "NewArrayEncodedMap should create an empty backing items map")
}

func TestArrayEncodedMap(t *testing.T) {
	// Validate that we can work with the data like a map
	wantItems := map[string]TestMapEntry{
		"x": {Value: "foo"},
		"y": {Value: "bar"},
	}

	// initialize the map
	m := NewArrayEncodedMap[TestMapEntry, TestArrayEntry]()
	for k, v := range wantItems {
		m.Set(k, v)
	}

	// make sure the data was persisted and can be retrieved
	assert.Equal(t, wantItems, m.items, "incorrect backing items persisted")
	assert.Equal(t, m.items, m.Items(), "incorrect Items() returned")

	// iterate the sorted items
	wantSorted := []TestArrayEntry{
		{Name: "x", Value: "foo"},
		{Name: "y", Value: "bar"},
	}
	assert.Equal(t, wantSorted, m.ItemsSorted(), "incorrect ItemsSorted() returned")

	// Get a specific item
	gotX, ok := m.Get("x")
	require.True(t, ok, "Get did not find x")
	assert.Equal(t, wantItems["x"], gotX, "incorrect x item retrieved")

	// Remove an item
	m.Remove("y")

	// Get the removed item
	_, ok = m.Get("y")
	require.False(t, ok, "Get should not have found 'y' because it was removed")
}

func TestArrayEncodedMap_ItemsUnsafe(t *testing.T) {
	t.Run("initialized", func(t *testing.T) {
		m := NewArrayEncodedMap[TestMapEntry, TestArrayEntry]()

		// Check that they reference the same map
		backingItems := m.items
		gotItemsUnsafe := *m.ItemsUnsafe()
		assert.Equal(t, reflect.ValueOf(backingItems).Pointer(), reflect.ValueOf(gotItemsUnsafe).Pointer(), "expected ItemsUnsafe to return the underlying map")
	})

	t.Run("uninitialized", func(t *testing.T) {
		var m ArrayEncodedMap[TestMapEntry, TestArrayEntry]
		assert.NotNil(t, m.ItemsUnsafe(), "expected ItemsUnsafe to not blow up when ArrayEncodedMap is uninitialized")
		assert.NotNil(t, m.items, "expected the backing items to be initialized on access when possible")

		// They should still reference the same map
		backingItems := m.items
		gotItemsUnsafe := *m.ItemsUnsafe()
		assert.Equal(t, reflect.ValueOf(backingItems).Pointer(), reflect.ValueOf(gotItemsUnsafe).Pointer(), "expected ItemsUnsafe to return the underlying map")
	})

	t.Run("nil", func(t *testing.T) {
		var m *ArrayEncodedMap[TestMapEntry, TestArrayEntry]
		assert.Nil(t, m.ItemsUnsafe(), "expected ItemsUnsafe to not blow up when ArrayEncodedMap is nil")
	})
}

func TestArrayEncodedMap_Merge(t *testing.T) {
	m := &ArrayEncodedMap[TestMapEntry, TestArrayEntry]{}
	m.Set("first", TestMapEntry{Value: "base first"})
	m.Set("second", TestMapEntry{Value: "base second"})
	m.Set("third", TestMapEntry{Value: "base third"})

	result := m.Merge(&ArrayEncodedMap[TestMapEntry, TestArrayEntry]{})
	require.Equal(t, 3, result.Len())
	_, ok := result.Get("fourth")
	assert.False(t, ok, "fourth should not be present in the base set")

	wantFourth := TestMapEntry{Value: "new fourth"}
	fourthMap := &ArrayEncodedMap[TestMapEntry, TestArrayEntry]{items: map[string]TestMapEntry{"fourth": wantFourth}}
	result = m.Merge(fourthMap)
	require.Equal(t, 4, result.Len())
	gotFourth, ok := result.Get("fourth")
	require.True(t, ok, "fourth should be present in the merged set")
	assert.Equal(t, wantFourth, gotFourth, "incorrect merged value for fourth")

	wantSecond := TestMapEntry{Value: "override second"}
	secondMap := &ArrayEncodedMap[TestMapEntry, TestArrayEntry]{items: map[string]TestMapEntry{"second": wantSecond}}
	result = m.Merge(secondMap)
	require.Equal(t, 3, result.Len())
	gotSecond, ok := result.Get("second")
	require.True(t, ok, "second should be present in the merged set")
	assert.Equal(t, wantSecond, gotSecond, "incorrect merged value for second")
}

func TestArrayEncodedMap_RoundTripMarshal(t *testing.T) {
	testcases := []struct {
		encoding string
	}{
		{encoding: "json"},
		{encoding: "yaml"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.encoding, func(t *testing.T) {
			testFile := fmt.Sprintf("testdata/array.%s", tc.encoding)

			var m *ArrayEncodedMap[TestMapEntry, TestArrayEntry]

			// Unmarshal
			fx := aferox.NewAferox(".", afero.NewOsFs())
			err := UnmarshalFile(fx, testFile, &m)
			require.NoError(t, err, "UnmarshalFile failed")

			// Validate the loaded data
			require.Equal(t, 2, m.Len(), "unexpected number of items defined")
			gotA, ok := m.Get("aname")
			require.True(t, ok, "aname was not defined")
			wantA := TestMapEntry{Value: "stuff"}
			assert.Equal(t, wantA, gotA, "unexpected aname defined")
			gotB, ok := m.Get("bname")
			require.True(t, ok, "password was not defined")
			wantB := TestMapEntry{Value: "things"}
			assert.Equal(t, wantB, gotB, "unexpected bname defined")

			// Marshal
			data, err := Marshal(tc.encoding, m)
			require.NoError(t, err, "Marshal failed")
			test.CompareGoldenFile(t, testFile, string(data))
		})
	}
}

func TestArrayEncodedMap_UnmarshalIntoNil(t *testing.T) {
	testcases := []struct {
		encoding string
	}{
		{encoding: "json"},
		{encoding: "yaml"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.encoding, func(t *testing.T) {
			testFile := fmt.Sprintf("testdata/nested-array.%s", tc.encoding)

			var dest testMap

			// Unmarshal
			fx := aferox.NewAferox(".", afero.NewOsFs())
			err := UnmarshalFile(fx, testFile, &dest)
			require.NoError(t, err, "UnmarshalFile failed")

			assert.Equal(t, 2, dest.Items.Len())
		})
	}
}

func TestArrayEncodedMap_Unmarshal_DuplicateKeys(t *testing.T) {
	data, err := os.ReadFile("testdata/array-with-duplicates.yaml")
	require.NoError(t, err, "ReadFile failed")

	var l ArrayEncodedMap[TestMapEntry, TestArrayEntry]
	err = yaml.Unmarshal(data, &l)
	require.ErrorContains(t, err, "cannot unmarshal source map: duplicate key found 'aname'")
}

type testMap struct {
	Items *ArrayEncodedMap[TestMapEntry, TestArrayEntry] `json:"items"`
}

// check that when we marshal an empty or nil ArrayEncodedMap, it stays null and isn't initialized to an _empty_ ArrayEncodedMap
// This impacts how it is marshaled later to yaml or json, because we often have fields tagged with omitempty
// and so it must be null to not be written out.
func TestArrayEncodedMap_MarshalEmptyToNull(t *testing.T) {
	testcases := []struct {
		name string
		src  testMap
	}{
		{name: "nil", src: testMap{Items: nil}},
		{name: "empty", src: testMap{Items: &ArrayEncodedMap[TestMapEntry, TestArrayEntry]{}}},
	}

	wantData, err := os.ReadFile("testdata/array-empty.json")
	require.NoError(t, err, "ReadFile failed")

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gotData, err := json.Marshal(tc.src)
			require.NoError(t, err, "Marshal failed")
			require.Equal(t, string(wantData), string(gotData), "empty ArrayEncodedMap should not marshal as empty, but nil so that it works with omitempty")
		})
	}
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
