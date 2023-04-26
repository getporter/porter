package secrets

/*
func TestSourceMapList_ToResolvedValues(t *testing.T) {
	l := Map{
		"bar": ValueMapping{ResolvedValue: "2"},
		"foo": ValueMapping{ResolvedValue: "1"},
	}

	want := map[string]string{
		"bar": "2",
		"foo": "1",
	}
	got := l.ToResolvedValues()
	assert.Equal(t, want, got)
}

func TestSourceMapList_Merge(t *testing.T) {
	set := Map{
		"first": ValueMapping{
			Source:        Source{Strategy: "value", Hint: "base first"},
			ResolvedValue: "base first"},
		"second": ValueMapping{
			Source:        Source{Strategy: "value", Hint: "base second"},
			ResolvedValue: "base second"},
		"third": ValueMapping{
			Source:        Source{Strategy: "value", Hint: "base third"},
			ResolvedValue: "base third"},
	}

	is := assert.New(t)

	result := set.Merge(Map{})
	is.Len(result, 3)
	is.NotContains(result, "fourth", "fourth should not be present in the base set")

	wantFourth := ValueMapping{
		Source:        Source{Strategy: "env", Hint: "FOURTH"},
		ResolvedValue: "new fourth"}
	fourth := Map{"fourth": wantFourth}
	result = set.Merge(fourth)
	is.Len(result, 4)
	gotFourth, ok := result["fourth"]
	is.True(ok, "fourth should be present in the merged set")
	is.Equal(wantFourth, gotFourth, "incorrect merged value for fourth")

	wantSecond := ValueMapping{
		Source:        Source{Strategy: "env", Hint: "SECOND"},
		ResolvedValue: "override second"}
	result = set.Merge(Map{"second": wantSecond})
	is.Len(result, 3)
	gotSecond, ok := result["second"]
	is.True(ok, "second should be present in the merged set")
	is.Equal(wantSecond, gotSecond, "incorrect merged value for second")
}

func TestSourceMapList_Unmarshal(t *testing.T) {
	data, err := os.ReadFile("testdata/strategies.yaml")
	require.NoError(t, err, "ReadFile failed")

	var l Map
	err = yaml.Unmarshal(data, &l)
	require.NoError(t, err, "Unmarshal failed")

	require.Len(t, l, 2, "unexpected number of strategies defined")

	gotLogLevel, ok := l["logLevel"]
	require.True(t, ok, "logLevel was not defined")
	wantLogLevel := ValueMapping{
		Source: Source{
			Strategy: "env",
			Hint:     "LOG_LEVEL",
		},
	}
	assert.Equal(t, wantLogLevel, gotLogLevel, "unexpected logLevel defined")

	gotPassword, ok := l["password"]
	require.True(t, ok, "password was not defined")
	wantPassword := ValueMapping{
		Source: Source{
			Strategy: "secret",
			Hint:     "my-password",
		},
	}
	assert.Equal(t, wantPassword, gotPassword, "unexpected password defined")
}

func TestSourceMapList_Unmarshal_DuplicateKeys(t *testing.T) {
	data, err := os.ReadFile("testdata/duplicate-strategies.yaml")
	require.NoError(t, err, "ReadFile failed")

	var l Map
	err = yaml.Unmarshal(data, &l)
	require.ErrorContains(t, err, "cannot unmarshal source map: duplicate key found 'logLevel'")
}
*/
