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

package rolebinding

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	rbac "k8s.io/kubernetes/pkg/apis/rbac"
)

var apiVersions = []string{"v1", "v1alpha1", "v1beta1"}

func TestDeclarativeValidateForDeclarative(t *testing.T) {
	for _, apiVersion := range apiVersions {
		t.Run(apiVersion, func(t *testing.T) {
			testDeclarativeValidateForDeclarative(t, apiVersion)
		})
	}
}

func testDeclarativeValidateForDeclarative(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "rbac.authorization.k8s.io",
		APIVersion: apiVersion,
	})

	testCases := map[string]struct {
		input        rbac.RoleBinding
		expectedErrs field.ErrorList
	}{
		"missing roleRef.name": {
			input: rbac.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "test-binding", Namespace: "test-ns"},
				RoleRef: rbac.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					// Name is intentionally missing here
				},
				Subjects: []rbac.Subject{
					{Kind: rbac.UserKind, APIGroup: rbac.GroupName, Name: "user1"},
				},
			},
			expectedErrs: field.ErrorList{
				field.Required(field.NewPath("roleRef", "name"), "name is required"),
			},
		},

		"missing subject.name": {
			input: rbac.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "test-binding", Namespace: "test-ns"},
				RoleRef: rbac.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     "reader",
				},
				Subjects: []rbac.Subject{
					{Kind: rbac.UserKind, APIGroup: rbac.GroupName},
				},
			},
			expectedErrs: field.ErrorList{
				field.Required(field.NewPath("subjects").Index(0).Child("name"), "name is required"),
			},
		},

		"valid binding": {
			input: rbac.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "valid-binding", Namespace: "test-ns"},
				RoleRef: rbac.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     "admin",
				},
				Subjects: []rbac.Subject{
					{Kind: rbac.UserKind, APIGroup: rbac.GroupName, Name: "user1"},
				},
			},
			expectedErrs: field.ErrorList{},
		},
		// TODO: Add more test cases
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			apitesting.VerifyValidationEquivalence(t, ctx, &tc.input, Strategy.Validate, tc.expectedErrs)
		})
	}
}

func TestDeclarativeValidateUpdateForDeclarative(t *testing.T) {
	for _, apiVersion := range apiVersions {
		t.Run(apiVersion, func(t *testing.T) {
			testDeclarativeValidateUpdateForDeclarative(t, apiVersion)
		})
	}
}

func testDeclarativeValidateUpdateForDeclarative(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "rbac.authorization.k8s.io",
		APIVersion: apiVersion,
	})

	validRoleBinding := rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "test-binding", Namespace: "test-ns"},
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "reader",
		},
		Subjects: []rbac.Subject{
			{Kind: rbac.UserKind, APIGroup: rbac.GroupName, Name: "user1"},
		},
	}

	testCases := map[string]struct {
		newObj       rbac.RoleBinding
		oldObj       rbac.RoleBinding
		expectedErrs field.ErrorList
	}{
		"no change to roleRef": {
			newObj:       validRoleBinding,
			oldObj:       validRoleBinding,
			expectedErrs: field.ErrorList{},
		},
		"change roleRef.name": {
			newObj: func() rbac.RoleBinding {
				rb := validRoleBinding.DeepCopy()
				rb.RoleRef.Name = "writer"
				return *rb
			}(),
			oldObj: validRoleBinding,
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("roleRef"), nil, "").WithOrigin("immutable").MarkShortCircuitedInDV(),
			},
		},
		"change roleRef.kind": {
			newObj: func() rbac.RoleBinding {
				rb := validRoleBinding.DeepCopy()
				rb.RoleRef.Kind = "ClusterRole"
				return *rb
			}(),
			oldObj: validRoleBinding,
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("roleRef"), nil, "").WithOrigin("immutable").MarkShortCircuitedInDV(),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			apitesting.VerifyUpdateValidationEquivalence(t, ctx, &tc.newObj, &tc.oldObj, Strategy.ValidateUpdate, tc.expectedErrs)
		})
	}
}
