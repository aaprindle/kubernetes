/*
Copyright 2024 The Kubernetes Authors.

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

package minimum

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
)

func TestBasicStruct(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	st.Value(&BasicStruct{
		// all zero values
		IntPtrField:     ptr.To(0),
		UintPtrField:    ptr.To(uint(0)),
		TypedefPtrField: ptr.To(IntType(0)),
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField().ByDetailSubstring(), field.ErrorList{
		field.Invalid(field.NewPath("intField"), nil, ""),
		field.Invalid(field.NewPath("intPtrField"), nil, ""),
		field.Invalid(field.NewPath("int16Field"), nil, ""),
		field.Invalid(field.NewPath("int32Field"), nil, ""),
		field.Invalid(field.NewPath("int64Field"), nil, ""),
		field.Invalid(field.NewPath("uintField"), nil, ""),
		field.Invalid(field.NewPath("uintPtrField"), nil, ""),
		field.Invalid(field.NewPath("uint16Field"), nil, ""),
		field.Invalid(field.NewPath("uint32Field"), nil, ""),
		field.Invalid(field.NewPath("uint64Field"), nil, ""),
		field.Invalid(field.NewPath("typedefField"), nil, ""),
		field.Invalid(field.NewPath("typedefPtrField"), nil, ""),
	})
	// Test validation ratcheting
	st.Value(&BasicStruct{
		IntPtrField:     ptr.To(0),
		UintPtrField:    ptr.To(uint(0)),
		TypedefPtrField: ptr.To(IntType(0)),
	}).OldValue(&BasicStruct{
		IntPtrField:     ptr.To(0),
		UintPtrField:    ptr.To(uint(0)),
		TypedefPtrField: ptr.To(IntType(0)),
	}).ExpectValid()

	st.Value(&BasicStruct{
		IntField:        1,
		IntPtrField:     ptr.To(1),
		Int16Field:      1,
		Int32Field:      1,
		Int64Field:      1,
		UintField:       1,
		Uint16Field:     1,
		Uint32Field:     1,
		Uint64Field:     1,
		UintPtrField:    ptr.To(uint(1)),
		TypedefField:    IntType(1),
		TypedefPtrField: ptr.To(IntType(1)),
	}).ExpectValid()
}

func TestOptionalStruct(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Zero values should be valid for optional fields
	st.Value(&OptionalStruct{}).ExpectValid()

	// Non-nil pointer with zero value should fail minimum
	st.Value(&OptionalStruct{
		OptionalIntPtrField: ptr.To(0),
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField().ByOrigin(), field.ErrorList{
		field.Invalid(field.NewPath("optionalIntPtrField"), nil, ""),
	})

	// Valid values
	st.Value(&OptionalStruct{
		OptionalIntField:    1,
		OptionalIntPtrField: ptr.To(1),
	}).ExpectValid()
}

func TestRequiredStruct(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Zero values should fail for required fields
	st.Value(&RequiredStruct{}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField().ByOrigin(), field.ErrorList{
		field.Invalid(field.NewPath("requiredIntField"), nil, ""),
		field.Required(field.NewPath("requiredIntPtrField"), ""),
	})

	// Valid values
	st.Value(&RequiredStruct{
		RequiredIntField:    1,
		RequiredIntPtrField: ptr.To(1),
	}).ExpectValid()

	// Test validation ratcheting
	st.Value(&RequiredStruct{}).OldValue(&RequiredStruct{}).ExpectValid()
}

func TestNegativeMinimumStruct(t *testing.T) {
	st := localSchemeBuilder.Test(t)

	// Zero values are above -10, should be valid
	st.Value(&NegativeMinimumStruct{}).ExpectValid()

	// Values below -10 should fail
	st.Value(&NegativeMinimumStruct{
		NegativeMinimumField:            -11,
		NegativeMinimumPtrField:         ptr.To(-11),
		OptionalNegativeMinimumField:    -11,
		OptionalNegativeMinimumPtrField: ptr.To(-11),
		RequiredNegativeMinimumField:    -11,
		RequiredNegativeMinimumPtrField: ptr.To(-11),
	}).ExpectMatches(field.ErrorMatcher{}.ByType().ByField().ByOrigin(), field.ErrorList{
		field.Invalid(field.NewPath("negativeMinimumField"), nil, ""),
		field.Invalid(field.NewPath("negativeMinimumPtrField"), nil, ""),
		field.Invalid(field.NewPath("optionalNegativeMinimumField"), nil, ""),
		field.Invalid(field.NewPath("optionalNegativeMinimumPtrField"), nil, ""),
		field.Invalid(field.NewPath("requiredNegativeMinimumField"), nil, ""),
		field.Invalid(field.NewPath("requiredNegativeMinimumPtrField"), nil, ""),
	})

	// Values at exactly -10 should be valid
	st.Value(&NegativeMinimumStruct{
		NegativeMinimumField:            -10,
		NegativeMinimumPtrField:         ptr.To(-10),
		OptionalNegativeMinimumField:    -10,
		OptionalNegativeMinimumPtrField: ptr.To(-10),
		RequiredNegativeMinimumField:    -10,
		RequiredNegativeMinimumPtrField: ptr.To(-10),
	}).ExpectValid()

	// Test validation ratcheting
	st.Value(&NegativeMinimumStruct{
		NegativeMinimumField:    -11,
		NegativeMinimumPtrField: ptr.To(-11),
	}).OldValue(&NegativeMinimumStruct{
		NegativeMinimumField:    -11,
		NegativeMinimumPtrField: ptr.To(-11),
	}).ExpectValid()
}
