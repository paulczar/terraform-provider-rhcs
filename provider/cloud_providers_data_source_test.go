/*
Copyright (c) 2021 Red Hat, Inc.

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

package provider

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"                         // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

var _ = Describe("Cloud providers data source", func() {
	var ctx context.Context
	var server *Server
	var ca string
	var token string

	BeforeEach(func() {
		// Create a contet:
		ctx = context.Background()

		// Create an access token:
		token = MakeTokenString("Bearer", 10*time.Minute)

		// Start the server:
		server, ca = MakeTCPTLSServer()
	})

	AfterEach(func() {
		// Stop the server:
		server.Close()

		// Remove the server CA file:
		err := os.Remove(ca)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Can list cloud providers", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "aws",
				      "name": "aws",
				      "display_name": "AWS"
				    },
				    {
				      "id": "gcp",
				      "name": "gcp",
				      "display_name": "GCP"
				    }
				  ]
				}`),
			),
		)

		// Run the apply command:
		result := NewTerraformRunner().
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source = "localhost/openshift-online/ocm"
				    }
				  }
				}

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}")
				}

				data "ocm_cloud_providers" "all" {
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Apply(ctx)
		Expect(result.ExitCode()).To(BeZero())

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "all")
		Expect(resource).To(MatchJQ(`.attributes.items | length`, 2))
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "aws"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].name`, "aws"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].display_name`, "AWS"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].id`, "gcp"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].name`, "gcp"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].display_name`, "GCP"))
	})

	It("Can search cloud providers", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"),
				VerifyFormKV("search", "display_name like 'A%'"),
				VerifyFormKV("order", "display_name asc"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "aws",
				      "name": "aws",
				      "display_name": "AWS"
				    },
				    {
				      "id": "azure",
				      "name": "azure",
				      "display_name": "Azure"
				    }
				  ]
				}`),
			),
		)

		// Run the apply command:
		result := NewTerraformRunner().
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source = "localhost/openshift-online/ocm"
				    }
				  }
				}

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}")
				}

				data "ocm_cloud_providers" "a" {
				  search = "display_name like 'A%'"
				  order  = "display_name asc"
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Apply(ctx)
		Expect(result.ExitCode()).To(BeZero())

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "a")
		Expect(resource).To(MatchJQ(`.attributes.search`, "display_name like 'A%'"))
		Expect(resource).To(MatchJQ(`.attributes.order`, "display_name asc"))
		Expect(resource).To(MatchJQ(`.attributes.items | length`, 2))
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "aws"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].name`, "aws"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].display_name`, "AWS"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].id`, "azure"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].name`, "azure"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].display_name`, "Azure"))
	})

	It("Populates `item` if there is exactly one result", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 1,
				  "total": 1,
				  "items": [
				    {
				      "id": "aws",
				      "name": "aws",
				      "display_name": "AWS"
				    }
				  ]
				}`),
			),
		)

		// Run the apply command:
		result := NewTerraformRunner().
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source = "localhost/openshift-online/ocm"
				    }
				  }
				}

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}")
				}

				data "ocm_cloud_providers" "a" {
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Apply(ctx)
		Expect(result.ExitCode()).To(BeZero())

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "a")
		Expect(resource).To(MatchJQ(`.attributes.item.id`, "aws"))
		Expect(resource).To(MatchJQ(`.attributes.item.name`, "aws"))
		Expect(resource).To(MatchJQ(`.attributes.item.display_name`, "AWS"))
	})

	It("Doesn't populate `item` if there are zero results", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 0,
				  "total": 0,
				  "items": []
				}`),
			),
		)

		// Run the apply command:
		result := NewTerraformRunner().
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source = "localhost/openshift-online/ocm"
				    }
				  }
				}

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}")
				}

				data "ocm_cloud_providers" "all" {
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Apply(ctx)
		Expect(result.ExitCode()).To(BeZero())

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "all")
		Expect(resource).To(MatchJQ(`.attributes.item`, nil))
	})

	It("Doesn't populate `item` if there multiple results", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/cloud_providers"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "aws",
				      "name": "aws",
				      "display_name": "AWS"
				    },
				    {
				      "id": "gcp",
				      "name": "gcp",
				      "display_name": "GCP"
				    }
				  ]
				}`),
			),
		)

		// Run the apply command:
		result := NewTerraformRunner().
			File(
				"main.tf", `
				terraform {
				  required_providers {
				    ocm = {
				      source = "localhost/openshift-online/ocm"
				    }
				  }
				}

				provider "ocm" {
				  url         = "{{ .URL }}"
				  token       = "{{ .Token }}"
				  trusted_cas = file("{{ .CA }}")
				}

				data "ocm_cloud_providers" "all" {
				}
				`,
				"URL", server.URL(),
				"Token", token,
				"CA", strings.ReplaceAll(ca, "\\", "/"),
			).
			Apply(ctx)
		Expect(result.ExitCode()).To(BeZero())

		// Check the state:
		resource := result.Resource("ocm_cloud_providers", "all")
		Expect(resource).To(MatchJQ(`.attributes.item`, nil))
	})
})
