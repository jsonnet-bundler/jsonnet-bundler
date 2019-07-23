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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/pkg/errors"
)

type GitPackage struct {
	Source *spec.GitSource
}

func NewGitPackage(source *spec.GitSource) Interface {
	return &GitPackage{
		Source: source,
	}
}

func (p *GitPackage) Install(ctx context.Context, name, dir, version string) (string, error) {
	destPath := path.Join(dir, name)

	tmpDir, err := ioutil.TempDir(filepath.Join(dir, ".tmp"), fmt.Sprintf("jsonnetpkg-%s-%s", name, version))
	if err != nil {
		return "", errors.Wrap(err, "failed to create tmp dir")
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.CommandContext(ctx, "git", "clone", p.Source.Remote, tmpDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	cmd = exec.CommandContext(ctx, "git", "-c", "advice.detachedHead=false", "checkout", version)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	b := bytes.NewBuffer(nil)
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Stdout = b
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	commitHash := strings.TrimSpace(b.String())

	err = os.RemoveAll(path.Join(tmpDir, ".git"))
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(path.Dir(destPath), os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "failed to create parent path")
	}

	err = os.RemoveAll(destPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to clean previous destination path")
	}

	err = os.Rename(path.Join(tmpDir, p.Source.Subdir), destPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to move package")
	}

	return commitHash, nil
}
