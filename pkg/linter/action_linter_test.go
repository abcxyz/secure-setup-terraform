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

func TestActionLinter_FindViolations(t *testing.T) {
	t.Parallel()

	withoutSetupTerraform := `
name: 'test-without-setup-tf'

on:
  pull_request:

permissions:
  contents: 'read'
  id-token: 'write'

jobs:
  dosomework:
    runs-on: 'ubuntu-latest'
    steps:
      - name: 'Checkout'
        uses: 'actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b' # ratchet:actions/checkout@v3

      - name: 'Authenticate to Google Cloud'
        uses: 'google-github-actions/auth@ceee102ec2387dd9e844e01b530ccd4ec87ce955' # ratchet:google-github-actions/auth@v0
        with:
          workload_identity_provider: 'projects/${{ env.PROJECT_NUMBER }}/locations/global/workloadIdentityPools/github-pool/providers/github-provider'
          service_account: 'gh-access-sa@lumberjack-dev-infra.iam.gserviceaccount.com'

      - name: 'Install and configure gcloud'
        uses: 'google-github-actions/setup-gcloud@877d4953d2c70a0ba7ef3290ae968eb24af233bb' # ratchet:google-github-actions/setup-gcloud@v0
 `
	withSetupTerraform := `
name: 'test-with-setup-tf'

on:
  pull_request:

permissions:
  contents: 'read'
  id-token: 'write'

jobs:
  dosomework:
    runs-on: 'ubuntu-latest'
    steps:
      - name: 'Checkout'
        uses: 'actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b' # ratchet:actions/checkout@v3

      - name: 'Authenticate to Google Cloud'
        uses: 'google-github-actions/auth@ceee102ec2387dd9e844e01b530ccd4ec87ce955' # ratchet:google-github-actions/auth@v0
        with:
          workload_identity_provider: 'projects/${{ env.PROJECT_NUMBER }}/locations/global/workloadIdentityPools/github-pool/providers/github-provider'
          service_account: 'gh-access-sa@lumberjack-dev-infra.iam.gserviceaccount.com'

      - name: 'Install and configure gcloud'
        uses: 'google-github-actions/setup-gcloud@877d4953d2c70a0ba7ef3290ae968eb24af233bb' # ratchet:google-github-actions/setup-gcloud@v0

  someotherjob:
    runs-on: 'ubuntu-latest'
    steps:
      - name: 'Setup Terraform'
        uses: 'hashicorp/setup-terraform@17d4c9b8043b238f6f35641cdd8433da1e6f3867' # ratchet:hashicorp/setup-terraform@v2
        with:
          terraform_wrapper: false
      - name: 'Init the terraform infrastructure'
        run: terraform -chdir=${{ env.tf_module_dir }} init
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
			name:        "yaml without setup-terraform action",
			filename:    "/test/myfile1",
			content:     withoutSetupTerraform,
			expectCount: 0,
			expect:      nil,
			wantError:   false,
		},
		{
			name:        "yaml with setup-terraform action",
			filename:    "/test/myfile2",
			content:     withSetupTerraform,
			expectCount: 1,
			expect: []*ViolationInstance{
				{
					ViolationType: "setup-terraform",
					Path:          "/test/myfile2",
					Line:          31,
				},
			},
			wantError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := GitHubActionLinter{}
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
