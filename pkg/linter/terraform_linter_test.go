// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package linter

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTerraformLinter_FindViolations(t *testing.T) {
	t.Parallel()

	withLocalExec := `
	resource "google_project_service" "run_api" {
		service = "run.googleapis.com"
		disable_on_destroy = true
	}
	resource "google_cloud_run_service" "run_service" {
		name     = var.service_name
		location = var.region
		template {
			spec {
				containers {
					image = var.service_image
				}
			}
		}
		traffic {
			percent         = 100
			latest_revision = true
		}
		depends_on = [google_project_service.run_api, null_resource.echo]
	}
	resource "null_resource" "echo" {
		provisioner "local-exec" {
			command = "echo this is a bad practice"
		}
	}
	`
	withoutLocalExec := `
	resource "google_project_service" "run_api" {
		service = "run.googleapis.com"
		disable_on_destroy = true
	}
	resource "google_cloud_run_service" "run_service" {
		name     = var.service_name
		location = var.region
		template {
			spec {
				containers {
					image = var.service_image
				}
			}
		}
		traffic {
			percent         = 100
			latest_revision = true
		}
		depends_on = [google_project_service.run_api]
	}
	`
	withLocalExecAsString := `
	resource "google_project_service" "local-exec" {
		service = "run.googleapis.com"
		disable_on_destroy = true
	}
	resource "google_cloud_run_service" "run_service" {
		name     = var.service_name
		location = var.region
		template {
			spec {
				containers {
					image = var.service_image
				}
			}
		}
		traffic {
			percent         = 100
			latest_revision = true
		}
		depends_on = [google_project_service.run_api]
	}
	`

	withRemoteExec := `
	resource "google_project_service" "run_api" {
		service = "run.googleapis.com"
		disable_on_destroy = true
	}
	resource "google_cloud_run_service" "run_service" {
		name     = var.service_name
		location = var.region
		template {
			spec {
				containers {
					image = var.service_image
				}
			}
		}
		traffic {
			percent         = 100
			latest_revision = true
		}
		depends_on = [google_project_service.run_api, null_resource.echo]
	}
	resource "null_resource" "echo" {
		provisioner "remote-exec" {
			inline = [
				"puppet apply",
				"consul join ${gcp_instance.web.private_ip}",
			  ]
		}
	}
	`

	cases := []struct {
		name        string
		filename    string
		content     string
		expectCount int
		expect      []*ViolationInstance
		wantError   bool
	}{
		{
			name:        "with local exec",
			filename:    "/my/path/to/testfile1",
			content:     withLocalExec,
			expectCount: 1,
			expect: []*ViolationInstance{
				{
					ViolationType: "local-exec",
					Path:          "/my/path/to/testfile1",
					Line:          23,
				},
			},
			wantError: false,
		},
		{
			name:        "without local exec",
			filename:    "/my/path/to/testfile2",
			content:     withoutLocalExec,
			expectCount: 0,
			expect:      nil,
			wantError:   false,
		},
		{
			name:        "with local exec as string",
			filename:    "/my/path/to/testfile3",
			content:     withLocalExecAsString,
			expectCount: 0,
			expect:      nil,
			wantError:   false,
		},
		{
			name:        "with remote exec",
			filename:    "/my/path/to/testfile1",
			content:     withRemoteExec,
			expectCount: 1,
			expect: []*ViolationInstance{
				{
					ViolationType: "remote-exec",
					Path:          "/my/path/to/testfile1",
					Line:          23,
				},
			},
			wantError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := TerraformLinter{}
			results, err := l.FindViolations([]byte(tc.content), tc.filename)
			if tc.wantError != (err != nil) {
				t.Errorf("expected error want: %#v, got: %#v - error: %v", tc.wantError, err != nil, err)
			}
			if diff := cmp.Diff(tc.expect, results); diff != "" {
				t.Errorf("results (-want,+got):\n%s", diff)
			}
		})
	}
}
