/*
Copyright 2022 The Kubernetes Authors.

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

// conformance_test contains code to run the conformance tests. This is in its own package to avoid circular imports.
package conformance_test

import (
	"strings"
	"testing"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	testconfig "sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

var testsToSkip = []string{
	// Expect 500 response status code which we cannot provide
	tests.HTTPRouteInvalidCrossNamespaceBackendRef.ShortName,
	tests.HTTPRouteInvalidBackendRefUnknownKind.ShortName,
	tests.HTTPRouteInvalidNonExistentBackendRef.ShortName,

	// Test asserts 404 response which we can't yet provide due to xDS control
	tests.HTTPRouteHeaderMatching.ShortName,
	tests.HTTPExactPathMatching.ShortName,

	// Tests create a gateway which gets stuck in status Unknown, with
	// reason NotReconciled, "Waiting for controller" (why?)
	tests.HTTPRouteListenerHostnameMatching.ShortName,
	tests.HTTPRouteDisallowedKind.ShortName,
	tests.HTTPRouteHostnameIntersection.ShortName,
}

func TestConformance(t *testing.T) {
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Error loading Kubernetes config: %v", err)
	}
	client, err := client.New(cfg, client.Options{})
	if err != nil {
		t.Fatalf("Error initializing Kubernetes client: %v", err)
	}
	v1alpha2.AddToScheme(client.Scheme())
	v1beta1.AddToScheme(client.Scheme())

	t.Logf("Running conformance tests with %s GatewayClass", *flags.GatewayClassName)

	supportedFeatures := []suite.SupportedFeature{}
	exemptFeatures := parseExemptFeatures(*flags.ExemptFeatures)

	cSuite := suite.New(suite.Options{
		Client:               client,
		GatewayClassName:     *flags.GatewayClassName,
		Debug:                *flags.ShowDebug,
		CleanupBaseResources: *flags.CleanupBaseResources,
		SupportedFeatures:    supportedFeatures,
		ExemptFeatures:       exemptFeatures,
		TimeoutConfig: testconfig.TimeoutConfig{
			MaxTimeToConsistency: 60 * time.Second,
		},
	})

	var testsToRun []suite.ConformanceTest
	for _, conformanceTest := range tests.ConformanceTests {
		if !slices.Contains(testsToSkip, conformanceTest.ShortName) {
			testsToRun = append(testsToRun, conformanceTest)
		}
	}

	cSuite.Setup(t)
	cSuite.Run(t, testsToRun)
}

// parseSupportedFeatures parses the arguments for supported-features flag,
// then converts the string to []suite.SupportedFeature
func parseSupportedFeatures(f string) []suite.SupportedFeature {
	var res []suite.SupportedFeature
	for _, value := range strings.Split(f, ",") {
		res = append(res, suite.SupportedFeature(value))
	}
	return res
}

// parseExemptFeatures parses the arguments for exempt-features flag,
// then converts the string to []suite.ExemptFeature
func parseExemptFeatures(f string) []suite.ExemptFeature {
	var res []suite.ExemptFeature
	for _, value := range strings.Split(f, ",") {
		res = append(res, suite.ExemptFeature(value))
	}
	return res
}
