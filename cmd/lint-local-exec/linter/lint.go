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

	"github.com/bradegler/secure-setup-terraform/cmd/lint-local-exec/version"
	"github.com/bradegler/secure-setup-terraform/pkg/lint"
)

const violationType = "local-exec"

var selectors = []string{"*.tf"}

type TerraformLinter struct{}

// FindViolations inspects a set of bytes that represent hcl from a terraform configuration file
// looking for calls to the 'local-exec' provider.
func (tfl *TerraformLinter) FindViolations(content []byte, path string) ([]lint.ViolationInstance, error) {
	tokens, diags := hclsyntax.LexConfig(content, path, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("error lexing hcl file contents: [%s]", diags.Error())
	}
	instances := []lint.ViolationInstance{}
	for _, token := range tokens {
		if token.Type == hclsyntax.TokenQuotedLit && string(token.Bytes) == violationType {
			instances = append(instances, lint.ViolationInstance{Path: path, Line: token.Range.Start.Line})
		}
	}
	return instances, nil
}

func (tfl *TerraformLinter) Selectors() []string   { return selectors }
func (tfl *TerraformLinter) ViolationType() string { return violationType }
func (tfl *TerraformLinter) Version() string       { return version.HumanVersion }
