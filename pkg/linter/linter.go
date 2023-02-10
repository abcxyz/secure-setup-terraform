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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ViolationInstance is an object that contains a reference to a location
// in a file where a lint violation was detected.
type ViolationInstance struct {
	ViolationType string
	Path          string
	Line          int
}

// Linter defines an interface selecting a set of files to apply lint rules
// against.
type Linter interface {
	// Selectors provides a set of file suffixes to search for. '.tf', '.yml', etc.
	Selectors() []string

	// FindViolations is the specific linter implementation that is applied to each
	// file to find any lint violations.
	FindViolations(content []byte, path string) ([]*ViolationInstance, error)
}

// RunLinter run executes the linter for a set of files.
func RunLinter(ctx context.Context, paths []string, linter Linter) error {
	var violations []*ViolationInstance
	// Process each provided path looking for violations
	for _, path := range paths {
		instances, err := lint(path, linter)
		if err != nil {
			return fmt.Errorf("error linting files: %w", err)
		}
		violations = append(violations, instances...)
	}
	for _, instance := range violations {
		fmt.Printf("%q detected at [%s:%d]\n", instance.ViolationType, instance.Path, instance.Line)
	}
	if len(violations) != 0 {
		return fmt.Errorf("found %d violation(s)", len(violations))
	}

	return nil
}

func lint(path string, linter Linter) ([]*ViolationInstance, error) {
	isDir, err := isDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file at path %q: %w", path, err)
	}
	instances := []*ViolationInstance{}
	if isDir {
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("error reading directory at path %q: %w", path, err)
		}
		for _, file := range files {
			next := filepath.Join(path, file.Name())
			results, err := lint(next, linter)
			if err != nil {
				return nil, err
			}
			if results != nil {
				instances = append(instances, results...)
			}
		}
	} else {
		for _, sel := range linter.Selectors() {
			if strings.HasSuffix(path, sel) {
				content, err := os.ReadFile(path)
				if err != nil {
					return nil, fmt.Errorf("error reading file: %w", err)
				}
				results, err := linter.FindViolations(content, path)
				if err != nil {
					return nil, fmt.Errorf("error searching for violations %w", err)
				}
				instances = append(instances, results...)
			}
		}
	}
	return instances, nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("error reading file information %w", err)
	}
	return fileInfo.IsDir(), err
}
