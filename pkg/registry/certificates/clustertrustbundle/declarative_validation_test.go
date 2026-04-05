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

package clustertrustbundle

import (
	"crypto/ed25519"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	mathrand "math/rand"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	"k8s.io/kubernetes/pkg/apis/certificates"
)

var apiVersions = []string{"v1alpha1", "v1beta1"}

func mustMakeCertificate(t *testing.T, template *x509.Certificate) []byte {
	gen := mathrand.New(mathrand.NewSource(12345))

	pub, priv, err := ed25519.GenerateKey(gen)
	if err != nil {
		t.Fatalf("Error while generating key: %v", err)
	}

	cert, err := x509.CreateCertificate(gen, template, template, pub, priv)
	if err != nil {
		t.Fatalf("Error while making certificate: %v", err)
	}

	return cert
}

func mustMakePEMBlock(data []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: data,
	}))
}

func makeGoodCertPEM(t *testing.T) string {
	cert := mustMakeCertificate(t, &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			CommonName: "root1",
		},
		IsCA:                  true,
		BasicConstraintsValid: true,
	})
	return mustMakePEMBlock(cert)
}

func makeValidBundle(t *testing.T, signerName string) certificates.ClusterTrustBundle {
	name := "test-bundle"
	if signerName != "" {
		// For signer-bound bundles, name must have the signer prefix (slashes to colons)
		name = "example.com:signer:abc"
		signerName = "example.com/signer"
	}
	return certificates.ClusterTrustBundle{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: certificates.ClusterTrustBundleSpec{
			SignerName:  signerName,
			TrustBundle: makeGoodCertPEM(t),
		},
	}
}

func TestDeclarativeValidateForDeclarative(t *testing.T) {
	for _, apiVersion := range apiVersions {
		testDeclarativeValidateForDeclarative(t, apiVersion)
	}
}

func testDeclarativeValidateForDeclarative(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "certificates.k8s.io",
		APIVersion: apiVersion,
	})
	testCases := map[string]struct {
		input        certificates.ClusterTrustBundle
		expectedErrs field.ErrorList
	}{
		"valid bundle without signer": {
			input: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "")
				return b
			}(),
		},
		"valid bundle with signer": {
			input: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "example.com/signer")
				return b
			}(),
		},
	}
	for k, tc := range testCases {
		t.Run(k, func(t *testing.T) {
			apitesting.VerifyValidationEquivalence(t, ctx, &tc.input, Strategy.Validate, tc.expectedErrs)
		})
	}
}

func TestValidateUpdateForDeclarative(t *testing.T) {
	for _, apiVersion := range apiVersions {
		testValidateUpdateForDeclarative(t, apiVersion)
	}
}

func testValidateUpdateForDeclarative(t *testing.T, apiVersion string) {
	ctx := genericapirequest.WithRequestInfo(genericapirequest.NewDefaultContext(), &genericapirequest.RequestInfo{
		APIGroup:   "certificates.k8s.io",
		APIVersion: apiVersion,
	})
	testCases := map[string]struct {
		old          certificates.ClusterTrustBundle
		update       certificates.ClusterTrustBundle
		expectedErrs field.ErrorList
	}{
		"signerName unchanged - valid": {
			old: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "example.com/signer")
				b.ResourceVersion = "1"
				return b
			}(),
			update: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "example.com/signer")
				b.ResourceVersion = "1"
				return b
			}(),
		},
		"signerName changed - invalid": {
			old: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "")
				b.Name = "test-bundle"
				b.Spec.SignerName = "example.com/old"
				b.ResourceVersion = "1"
				return b
			}(),
			update: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "")
				b.Name = "test-bundle"
				b.Spec.SignerName = "example.com/new"
				b.ResourceVersion = "1"
				return b
			}(),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("spec", "signerName"), nil, "").WithOrigin("immutable"),
			},
		},
		"signerName set from unset - invalid": {
			old: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "")
				b.ResourceVersion = "1"
				return b
			}(),
			update: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "")
				b.Spec.SignerName = "example.com/signer"
				b.ResourceVersion = "1"
				return b
			}(),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("spec", "signerName"), nil, "").WithOrigin("immutable"),
			},
		},
		"signerName unset from set - invalid": {
			old: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "")
				b.Name = "test-bundle"
				b.Spec.SignerName = "example.com/signer"
				b.ResourceVersion = "1"
				return b
			}(),
			update: func() certificates.ClusterTrustBundle {
				b := makeValidBundle(t, "")
				b.Name = "test-bundle"
				b.Spec.SignerName = ""
				b.ResourceVersion = "1"
				return b
			}(),
			expectedErrs: field.ErrorList{
				field.Invalid(field.NewPath("spec", "signerName"), nil, "").WithOrigin("immutable"),
			},
		},
	}
	for k, tc := range testCases {
		t.Run(k, func(t *testing.T) {
			apitesting.VerifyUpdateValidationEquivalence(t, ctx, &tc.update, &tc.old, Strategy.ValidateUpdate, tc.expectedErrs)
		})
	}
}
