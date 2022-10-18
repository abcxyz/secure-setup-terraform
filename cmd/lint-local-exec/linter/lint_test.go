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
	"fmt"
	"reflect"
	"testing"

	"github.com/bradegler/secure-setup-terraform/lib"
)

func TestLint_FindViolations(t *testing.T) {
	t.Parallel()

	withLocalExec := "resource \"google_project_service\" \"run_api\" {  service = \"run.googleapis.com\"  disable_on_destroy = true}resource \"google_cloud_run_service\" \"run_service\" {  name     = var.service_name  location = var.region  template {    spec {      containers {        image = var.service_image      }    }  }  traffic {    percent         = 100    latest_revision = true  }  depends_on = [google_project_service.run_api, null_resource.echo]}resource \"null_resource\" \"echo\" {  provisioner \"local-exec\" {    command = \"echo this is a bad practice\"  }}"
	withoutLocalExec := "resource \"google_project_service\" \"run_api\" {  service = \"run.googleapis.com\"  disable_on_destroy = true}resource \"google_cloud_run_service\" \"run_service\" {  name     = var.service_name  location = var.region  template {    spec {      containers {        image = var.service_image      }    }  }  traffic {    percent         = 100    latest_revision = true  }  depends_on = [google_project_service.run_api]"

	cases := []struct {
		name        string
		filename    string
		content     string
		expectCount int
		expect      []lib.ViolationInstance
		wantError   bool
	}{
		{name: "with local exec", filename: "/my/path/to/testfile1", content: withLocalExec, expectCount: 1, expect: []lib.ViolationInstance{
			{Path: "/my/path/to/testfile1", Line: 1},
		}, wantError: false},
		{name: "without local exec", filename: "/my/path/to/testfile2", content: withoutLocalExec, expectCount: 0, expect: nil, wantError: false},
	}

	for _, tc := range cases {
		tc := tc // IMPORTANT: don't shadow the test case

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fmt.Printf("Test: %s\n", tc.name)

			l := TerraformLinter{}
			results, err := l.FindViolations([]byte(tc.content), tc.filename)
			if tc.wantError != (err != nil) {
				t.Fatalf("expected error: %#v, got: %#v - %v", tc.wantError, err != nil, err)
			}
			if tc.expectCount != len(results) {
				t.Fatalf("execpted results: %d, got: %d", tc.expectCount, len(results))
			}
			if len(tc.expect) != 0 && !reflect.DeepEqual(results, tc.expect) {
				t.Fatalf("execpted results did not match: %v, got: %v", tc.expect, results)
			}
		})
	}
}
