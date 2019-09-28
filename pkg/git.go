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
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
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

func downloadGitHubArchive(filepath string, url string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	color.Cyan("GET %s %d", url, resp.StatusCode)

	// GitHub conveniently uses the commit SHA1 at the ETag
	// signature for the archive. This is needed when doing `jb update`
	// to resolve a ref (ie. "master") to a commit SHA1 for the lock file
	etagValue := resp.Header.Get(http.CanonicalHeaderKey("ETag"))
	commitShaPattern, _ := regexp.Compile("^\"([0-9a-f]{40})\"$")
	m := commitShaPattern.FindStringSubmatch(etagValue)
	if len(m) < 2 {
		return "", errors.New(fmt.Sprintf("unexpected etag format: %s", etagValue))
	}
	commitSha := m[1]
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return commitSha, nil
}

func gzipUntar(dst string, r io.Reader, subDir string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		// strip the two first components of the path
		parts := strings.SplitAfterN(header.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		suffix := parts[1]
		prefix := dst

		// reconstruct the target parh for the archive entry
		target := filepath.Join(prefix, suffix)

		// if subdir is provided and target is not under it, skip it
		subDirPath := filepath.Join(prefix, subDir)
		if subDir != "" && !strings.HasPrefix(target, subDirPath) {
			continue
		}

		// check the file type
		switch header.Typeflag {

		// create directories as needed
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			f.Close()
		}
	}
}

func (p *GitPackage) Install(ctx context.Context, name, dir, version string) (string, error) {
	destPath := path.Join(dir, name)

	tmpDir, err := ioutil.TempDir(filepath.Join(dir, ".tmp"), fmt.Sprintf("jsonnetpkg-%s-%s", name, version))
	if err != nil {
		return "", errors.Wrap(err, "failed to create tmp dir")
	}
	defer os.RemoveAll(tmpDir)

	// Optimization for GitHub sources: download a tarball archive of the requested
	// version instead of cloning the entire repository. Resolves the version to a
	// commit SHA using the GitHub API.
	if strings.HasPrefix(p.Source.Remote, "https://github.com/") {
		archiveUrl := fmt.Sprintf("%s/archive/%s.tar.gz", p.Source.Remote, version)
		archiveFilepath := fmt.Sprintf("%s.tar.gz", tmpDir)

		defer os.Remove(archiveFilepath)
		commitSha, err := downloadGitHubArchive(archiveFilepath, archiveUrl)
		if err != nil {
			return "", err
		}
		r, err := os.Open(archiveFilepath)
		err = gzipUntar(tmpDir, r, p.Source.Subdir)
		return commitSha, nil
	}

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	cmd = exec.CommandContext(ctx, "git", "remote", "add", "origin", p.Source.Remote)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// Attempt shallow fetch at specific revision
	cmd = exec.CommandContext(ctx, "git", "fetch", "--depth", "1", "origin", version)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = tmpDir
	err = cmd.Run()
	if err != nil {
		// Fall back to normal fetch (all revisions)
		cmd = exec.CommandContext(ctx, "git", "fetch", "origin")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = tmpDir
		err = cmd.Run()
		if err != nil {
			return "", err
		}
	}

	// Sparse checkout optimization: if a Subdir is specificied,
	// there is no need to do a full checkout
	if p.Source.Subdir != "" {
		cmd = exec.CommandContext(ctx, "git", "config", "core.sparsecheckout", "true")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = tmpDir
		err = cmd.Run()
		if err != nil {
			return "", err
		}
		glob := []byte(p.Source.Subdir + "/*\n")
		ioutil.WriteFile(tmpDir+"/.git/info/sparse-checkout", glob, 0644)
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
