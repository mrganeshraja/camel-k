// +build integration

// To enable compilation of this file in Goland, go to "Settings -> Go -> Vendoring & Build Tags -> Custom Tags" and add "integration"

/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/apache/camel-k/pkg/util/openshift"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestPlatformlessRun(t *testing.T) {
	needsStagingRepo := os.Getenv("STAGING_RUNTIME_REPO") != ""
	ocp, err := openshift.IsOpenShift(testClient)
	assert.Nil(t, err)
	if needsStagingRepo || !ocp {
		t.Skip("This test is for OpenShift only and cannot work when a custom platform configuration is needed")
		return
	}

	withNewTestNamespace(t, func(ns string) {
		Expect(kamel("install", "-n", ns).Execute()).Should(BeNil())

		// Delete the platform from the namespace before running the integration
		Eventually(deletePlatform(ns)).Should(BeTrue())

		Expect(kamel("run", "-n", ns, "files/yaml.yaml").Execute()).Should(BeNil())
		Eventually(integrationPodPhase(ns, "yaml"), 5*time.Minute).Should(Equal(v1.PodRunning))
		Eventually(integrationLogs(ns, "yaml"), 1*time.Minute).Should(ContainSubstring("Magicstring!"))

		// Platform should be recreated
		Eventually(platform(ns)).ShouldNot(BeNil())
		Expect(kamel("delete", "--all", "-n", ns).Execute()).Should(BeNil())
	})
}
