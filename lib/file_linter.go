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
