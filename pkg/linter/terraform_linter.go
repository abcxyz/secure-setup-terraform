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

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

const tokenLocalExec = "local-exec"
const tokenRemoteExec = "remote-exec"

var terraformSelectors = []string{".tf", ".tf.json"}

type TerraformLinter struct{}

// FindViolations inspects a set of bytes that represent hcl from a terraform configuration file
// looking for calls to the 'local-exec' provider.
func (tfl *TerraformLinter) FindViolations(content []byte, path string) ([]*ViolationInstance, error) {
	tokens, diags := hclsyntax.LexConfig(content, path, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("error lexing hcl file contents: [%s]", diags.Error())
	}

	var instances []*ViolationInstance
	inProvisioner := false
	for _, token := range tokens {
		if token.Bytes == nil {
			continue
		}

		// Each Ident token starts a new object, we are only looking for provisioner
		// type objects with specific types, local-exec and remote-exec
		if token.Type == hclsyntax.TokenIdent {
			inProvisioner = string(token.Bytes) == "provisioner"
		}
		if inProvisioner && token.Type == hclsyntax.TokenQuotedLit {
			if string(token.Bytes) == tokenLocalExec {
				instances = append(instances, &ViolationInstance{ViolationType: tokenLocalExec, Path: path, Line: token.Range.Start.Line})
			}
			if string(token.Bytes) == tokenRemoteExec {
				instances = append(instances, &ViolationInstance{ViolationType: tokenRemoteExec, Path: path, Line: token.Range.Start.Line})
			}
		}
	}
	return instances, nil
}

func (tfl *TerraformLinter) Selectors() []string { return terraformSelectors }
