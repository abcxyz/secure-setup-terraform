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

package lint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ViolationInstance is an object that contains a reference to a location
// in a file where a lint violation was detected.
type ViolationInstance struct {
	Path string
	Line int
}

// Linter defines an interface selecting a set of files to apply lint rules
// against.
type Linter interface {
	// ViolationType declears the type of violation the Linter identifies
	ViolationType() string
	// Version retrieves the human readable version string for the linter
	Version() string
	// Selectors provides a set of file suffixes to search for. '.tf', '.yml', etc.
	Selectors() []string
	// FindViolations is the specific linter implementation that is applied to each
	// file to find any lint violations.
	FindViolations(content []byte, path string) ([]ViolationInstance, error)
}

func lint(path string, linter Linter) ([]ViolationInstance, error) {
	isDir, err := isDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file at path %q: %w", path, err)
	}
	instances := []ViolationInstance{}
	if isDir {
		files, err := ioutil.ReadDir(path)
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
		return instances, err
	} else {
		for _, sel := range linter.Selectors() {
			if strings.HasSuffix(path, sel) {
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return nil, fmt.Errorf("error reading file: [%w]", err)
				}
				results, err := linter.FindViolations(content, path)
				if err != nil {
					return nil, err
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
		return false, err
	}
	return fileInfo.IsDir(), err
}
