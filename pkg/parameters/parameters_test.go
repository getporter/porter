/*
Copyright 2018 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This file has been modified from the original source
https://github.com/kubernetes-incubator/service-catalog/blob/129d98e6f6e3c65d16d47a26fcf3357f0c3478de/cmd/svcat/parameters/parameters_test.go
to use just one of the original functions that I (Carolyn Van Slyck) wrote for Service Catalog.
*/

package parameters

import (
	"reflect"
	"testing"
)

func TestParseVariableAssignments(t *testing.T) {
	testcases := []struct {
		Name, Raw, Variable, Value string
	}{
		{"simple", "a=b", "a", "b"},
		{"multiple equal signs", "c=abc1232===", "c", "abc1232==="},
		{"empty value", "d=", "d", ""},
		{"extra whitespace", " a = b ", "a", "b"},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {

			params := []string{tc.Raw}

			got, err := ParseVariableAssignments(params)
			if err != nil {
				t.Fatal(err)
			}

			want := make(map[string]string)
			want[tc.Variable] = tc.Value
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("%s\nexpected:\n\t%v\ngot:\n\t%v\n", tc.Raw, want, got)
			}
		})
	}
}

func TestParseVariableAssignments_MissingVariableName(t *testing.T) {
	params := []string{"=b"}

	_, err := ParseVariableAssignments(params)
	if err == nil {
		t.Fatal("should have failed due to a missing variable name")
	}
}
