package secrets

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrategyList_GetResolvedValue(t *testing.T) {
	l := SourceMapList{
		SourceMap{Name: "bar", ResolvedValue: "2"},
		SourceMap{Name: "foo", ResolvedValue: "1"},
	}

	fooVal, ok := l.GetResolvedValue("foo")
	require.True(t, ok, "foo could not be found")
	assert.Equal(t, "1", fooVal, "foo had the incorrect resolved value")

	_, ok = l.GetResolvedValue("missing")
	require.False(t, ok, "GetResolvedValue should have returned that missing key was not found")
}

func TestStrategyList_GetResolvedValues(t *testing.T) {
	l := SourceMapList{
		SourceMap{Name: "bar", ResolvedValue: "2"},
		SourceMap{Name: "foo", ResolvedValue: "1"},
	}

	want := map[string]string{
		"bar": "2",
		"foo": "1",
	}
	got := l.GetResolvedValues()
	assert.Equal(t, want, got)
}

func TestStrategyList_HasKey(t *testing.T) {
	l := SourceMapList{
		SourceMap{Name: "bar", ResolvedValue: "2"},
		SourceMap{Name: "foo", ResolvedValue: "1"},
	}

	require.True(t, l.HasName("foo"), "HasName returned the wrong value for a key present in the list")
	require.False(t, l.HasName("missing"), "HasName returned the wrong value for a missing key")
}

func TestStrategyList_Sort(t *testing.T) {
	l := SourceMapList{
		SourceMap{Name: "c"},
		SourceMap{Name: "a"},
		SourceMap{Name: "b"},
	}

	sort.Sort(l)

	wantResult := SourceMapList{
		SourceMap{Name: "a"},
		SourceMap{Name: "b"},
		SourceMap{Name: "c"},
	}

	require.Len(t, l, 3, "Len is not implemented correctly")
	require.Equal(t, wantResult, l, "Sort is not implemented correctly")
}

func TestStrategyList_GetStrategy(t *testing.T) {
	wantToken := SourceMap{Name: "token", Source: Source{Strategy: "env", Hint: "GITHUB_TOKEN"}}
	l := SourceMapList{
		SourceMap{Name: "logLevel", Source: Source{Strategy: "value", Hint: "11"}},
		wantToken,
	}

	gotToken, ok := l.GetByName("token")
	require.True(t, ok, "GetByName did not find 'token'")
	assert.Equal(t, wantToken, gotToken, "GetByName returned the wrong 'token' strategy")
}

func TestStrategyList_Merge(t *testing.T) {
	set := SourceMapList{
		SourceMap{Name: "first", ResolvedValue: "base first"},
		SourceMap{Name: "second", ResolvedValue: "base second"},
		SourceMap{Name: "third", ResolvedValue: "base third"},
	}

	is := assert.New(t)

	result := set.Merge(SourceMapList{})
	is.Len(result, 3)
	_, ok := result.GetResolvedValue("base fourth")
	is.False(ok, "forth should not be present in the base set")

	result = set.Merge(SourceMapList{SourceMap{Name: "fourth", ResolvedValue: "new fourth"}})
	is.Len(result, 4)
	val, ok := result.GetResolvedValue("fourth")
	is.True(ok, "fourth should be present when merged as an override")
	is.Equal("new fourth", val, "incorrect merged value")

	result = set.Merge(SourceMapList{SourceMap{Name: "second", ResolvedValue: "override second"}})
	is.Len(result, 3)
	val, ok = result.GetResolvedValue("second")
	is.True(ok, "second should be overwritten when an override value is merged")
	is.Equal("override second", val, "incorrect merged value")
}
