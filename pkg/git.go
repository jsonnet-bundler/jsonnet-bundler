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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/fatih/color"
)

type GitPackage struct {
	Source *spec.GitSource
}

func NewGitPackage(source *spec.GitSource) Interface {
	return &GitPackage{
		Source: source,
	}
}

func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (p *GitPackage) Install(ctx context.Context, dir, version string) (lockVersion string, err error) {
	if strings.HasPrefix(p.Source.Remote, "https://github.com/") {
		archiveUrl := fmt.Sprintf("%s/archive/%s.tar.gz", p.Source.Remote, version)
		archiveFilepath := fmt.Sprintf("%s.tar.gz", dir)
		err := DownloadFile(archiveFilepath, archiveUrl);
		if err != nil {
			return "", err;
		}
		color.Cyan("GET %s OK", archiveUrl);
		cmd := exec.CommandContext(ctx, "tar", "xvf", archiveFilepath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return "", err
		}
		color.Cyan("Untar %s OK", archiveFilepath);
		// TODO resolve git refs using GitHub API
		commitHash := version
		return commitHash, nil
	}

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	cmd = exec.CommandContext(ctx, "git", "remote", "add", "origin", p.Source.Remote)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// Attempt shallow fetch at specific revision
	cmd = exec.CommandContext(ctx, "git", "fetch", "--depth", "1", "origin", version)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	err = cmd.Run()
	if err != nil {
		// Fall back to normal fetch (all revisions)
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

	// If a Subdir is specificied, a sparsecheckout is sufficient
	if p.Source.Subdir != "" {
		cmd = exec.CommandContext(ctx, "git", "config", "core.sparsecheckout", "true")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		err = cmd.Run()
		if err != nil {
			return "", err
		}
		glob := []byte(p.Source.Subdir + "/*\n")
		ioutil.WriteFile(dir+"/.git/info/sparse-checkout", glob, 0644)
	}

	cmd = exec.CommandContext(ctx, "git", "-c", "advice.detachedHead=false", "checkout", version)
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

	err = os.RemoveAll(path.Join(dir, ".git"))
	if err != nil {
		return "", err
	}

	return commitHash, nil
}
