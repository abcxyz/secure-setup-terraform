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

package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ViolationInstance struct {
	Path string
	Line int
}

type Linter interface {
	GetViolationType() string
	GetVersion() string
	FindViolations(content []byte, path string) ([]ViolationInstance, error)
	GetSelectors() []string
}

type FileLinter interface {
	Lint(filePath string) ([]ViolationInstance, error)
}

type fileLinter struct {
	Linter Linter
}

func NewFileLinter(linter Linter) FileLinter {
	return &fileLinter{Linter: linter}
}

func (fl *fileLinter) Lint(path string) ([]ViolationInstance, error) {
	isDir, err := isDirectory(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file at path [%s]: %w", path, err)
	}
	if isDir {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("error reading directory at path [%s]: %w", path, err)
		}
		instances := []ViolationInstance{}
		for _, file := range files {
			next := filepath.Join(path, file.Name())
			results, err := fl.Lint(next)
			if err != nil {
				return nil, err
			}
			if results != nil {
				instances = append(instances, results...)
			}
		}
		return instances, err
	} else {
		for _, sel := range fl.Linter.GetSelectors() {
			if strings.HasSuffix(path, sel) {
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return nil, fmt.Errorf("error reading file: [%w]", err)
				}
				return fl.Linter.FindViolations(content, path)
			}
		}
	}
	return nil, nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}
