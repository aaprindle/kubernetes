/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mapofprimitive

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func Test(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&Struct{
		// All zero values
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.TooFew(field.NewPath("min10Field"), 0, 10),
		field.TooFew(field.NewPath("min10TypedefField"), 0, 10),
	})

	st.Value(&Struct{
		Min0Field:         make(map[string]string, 0),
		Min10Field:        make(map[string]string, 0),
		Min0TypedefField:  make(map[string]StringType, 0),
		Min10TypedefField: make(map[string]StringType, 0),
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.TooFew(field.NewPath("min10Field"), 0, 10),
		field.TooFew(field.NewPath("min10TypedefField"), 0, 10),
	})

	min10Field_1 := make(map[string]string)
	min10TypedefField_1 := make(map[string]StringType)
	for i := 0; i < 1; i++ {
		min10Field_1[fmt.Sprintf("k%d", i)] = "v"
		min10TypedefField_1[fmt.Sprintf("k%d", i)] = "v"
	}

	st.Value(&Struct{
		Min10Field:        min10Field_1,
		Min10TypedefField: min10TypedefField_1,
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.TooFew(field.NewPath("min10Field"), 1, 10),
		field.TooFew(field.NewPath("min10TypedefField"), 1, 10),
	})

	min10Field_9 := make(map[string]string)
	min10TypedefField_9 := make(map[string]StringType)
	for i := 0; i < 9; i++ {
		min10Field_9[fmt.Sprintf("k%d", i)] = "v"
		min10TypedefField_9[fmt.Sprintf("k%d", i)] = "v"
	}

	st.Value(&Struct{
		Min10Field:        min10Field_9,
		Min10TypedefField: min10TypedefField_9,
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField(), field.ErrorList{
		field.TooFew(field.NewPath("min10Field"), 9, 10),
		field.TooFew(field.NewPath("min10TypedefField"), 9, 10),
	})

	min10Field_10 := make(map[string]string)
	min10TypedefField_10 := make(map[string]StringType)
	for i := 0; i < 10; i++ {
		min10Field_10[fmt.Sprintf("k%d", i)] = "v"
		min10TypedefField_10[fmt.Sprintf("k%d", i)] = "v"
	}

	st.Value(&Struct{
		Min10Field:        min10Field_10,
		Min10TypedefField: min10TypedefField_10,
	}).ExpectValid()

	min0Field_1 := make(map[string]string)
	min10Field_11 := make(map[string]string)
	min0TypedefField_1 := make(map[string]StringType)
	min10TypedefField_11 := make(map[string]StringType)

	for i := 0; i < 1; i++ {
		min0Field_1[fmt.Sprintf("k%d", i)] = "v"
		min0TypedefField_1[fmt.Sprintf("k%d", i)] = "v"
	}
	for i := 0; i < 11; i++ {
		min10Field_11[fmt.Sprintf("k%d", i)] = "v"
		min10TypedefField_11[fmt.Sprintf("k%d", i)] = "v"
	}

	testVal := &Struct{
		Min0Field:         min0Field_1,
		Min10Field:        min10Field_11,
		Min0TypedefField:  min0TypedefField_1,
		Min10TypedefField: min10TypedefField_11,
	}
	st.Value(testVal).ExpectValid()

	// Test validation ratcheting
	st.Value(&Struct{
		Min0Field:         min0Field_1,
		Min10Field:        min10Field_1,
		Min0TypedefField:  min0TypedefField_1,
		Min10TypedefField: min10TypedefField_1,
	}).OldValue(&Struct{
		Min0Field:         min0Field_1,
		Min10Field:        min10Field_1,
		Min0TypedefField:  min0TypedefField_1,
		Min10TypedefField: min10TypedefField_1,
	}).ExpectValid()

}
