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
	"os"
	"os/exec"
	"strings"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

type GitPackage struct {
	Source *spec.GitSource
}

func NewGitPackage(source *spec.GitSource) Interface {
	return &GitPackage{
		Source: source,
	}
}

func (p *GitPackage) Install(ctx context.Context, dir, version string) (lockVersion string, err error) {
	_, err = os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		cmd := exec.CommandContext(ctx, "git", "clone", p.Source.Remote, dir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return "", err
		}
	} else {
		cmd := exec.CommandContext(ctx, "git", "remote", "set-url", "origin", p.Source.Remote)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		err = cmd.Run()
		if err != nil {
			return "", err
		}

		cmd = exec.CommandContext(ctx, "git", "fetch", "origin")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		err = cmd.Run()
		if err != nil {
			return "", err
		}
	}

	cmd := exec.CommandContext(ctx, "git", "checkout", "origin/"+version)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	b := bytes.NewBuffer(nil)
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Stdout = b
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	commitHash := strings.TrimSpace(b.String())
	return commitHash, nil
}
