// Copyright 2018 jsonnet-bundler authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

type LocalPackage struct {
	Source *spec.LocalSource
}

func NewLocalPackage(source *spec.LocalSource) Interface {
	return &LocalPackage{
		Source: source,
	}
}

func (p *LocalPackage) Install(ctx context.Context, name, dir, version string) (lockVersion string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	oldname := filepath.Join(wd, p.Source.Directory)
	newname := filepath.Join(wd, filepath.Join(dir, name))

	err = os.RemoveAll(newname)
	if err != nil {
		return "", fmt.Errorf("failed to clean previous destination path: %w", err)
	}

	_, err = os.Stat(oldname)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("symlink destination path does not exist: %w", err)
	}

	err = os.Symlink(oldname, newname)
	if err != nil {
		return "", fmt.Errorf("failed to create symlink for local dependency: %w", err)
	}

	return "", nil
}
