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
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

const tokenSetupTerraform = "setup-terraform"

var actionSelectors []string = []string{".yml", ".yaml"}

type GitHubActionLinter struct{}

// FindViolations inspects a set of bytes that represent a YAML document that defines
// a GitHub action workflow looking for steps that use the 'hashicorp/setup-terraform'
// action.
func (tfl *GitHubActionLinter) FindViolations(content []byte, path string) ([]*ViolationInstance, error) {
	reader := bytes.NewReader(content)
	node, err := parseYAML(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	if node == nil {
		return nil, nil
	}
	if node.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("expected document node, got %v", node.Kind)
	}

	var violations []*ViolationInstance
	// Top-level object map
	for _, docMap := range node.Content {
		if docMap.Kind != yaml.MappingNode {
			continue
		}
		// Top-level object map
		for _, docMap := range node.Content {
			if docMap.Kind != yaml.MappingNode {
				continue
			}

			for i, topLevelMap := range docMap.Content {
				// jobs: keyword
				if topLevelMap.Value == "jobs" {
					jobs := docMap.Content[i+1]
					if jobs.Kind != yaml.MappingNode {
						continue
					}

					for _, jobMap := range jobs.Content {
						if jobMap.Kind != yaml.MappingNode {
							continue
						}
						for j, sub := range jobMap.Content {

							// List of steps, iterate over each step and find the "uses" clause.
							if sub.Value == "steps" {
								steps := jobMap.Content[j+1]
								for _, step := range steps.Content {
									if step.Kind != yaml.MappingNode {
										continue
									}
									for k, property := range step.Content {
										if property.Value == "uses" {
											uses := step.Content[k+1]
											// Looking for the specific 'hashicorp/setup-terraform' action
											if strings.HasPrefix(uses.Value, "hashicorp/setup-terraform") {
												violations = append(violations, &ViolationInstance{ViolationType: tokenSetupTerraform, Path: path, Line: uses.Line})
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return violations, nil
}

func (tfl *GitHubActionLinter) Selectors() []string { return actionSelectors }

// parseYAML parses the given reader as a yaml node.
func parseYAML(r io.Reader) (*yaml.Node, error) {
	var m yaml.Node
	if err := yaml.NewDecoder(r).Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to decode yaml: %w", err)
	}
	return &m, nil
}
