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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

var (
	VersionMismatch = errors.New("multiple colliding versions specified")
)

func Ensure(want spec.JsonnetFile, vendorDir string, locks map[string]spec.Dependency) (map[string]spec.Dependency, error) {
	deps := make(map[string]spec.Dependency)

	for _, d := range want.Dependencies {
		l, present := locks[d.Name]

		// already locked and the integrity is intact
		if present && check(l, vendorDir) {
			deps[d.Name] = l
			continue
		}
		expectedSum := d.Sum

		// either not present or not intact: download again
		dir := filepath.Join(vendorDir, d.Name)
		os.RemoveAll(dir)

		locked, err := download(d, vendorDir)
		if err != nil {
			return nil, errors.Wrap(err, "downloading")
		}
		if expectedSum != "" && d.Sum != expectedSum {
			return nil, fmt.Errorf("checksum mismatch for %s. Expected %s but got %s", d.Name, expectedSum, d.Sum)
		}
		deps[d.Name] = *locked
	}

	for _, d := range deps {
		f, err := jsonnetfile.Load(filepath.Join(vendorDir, d.Name, jsonnetfile.File))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		nested, err := Ensure(f, vendorDir, locks)
		if err != nil {
			return nil, err
		}

		for _, d := range nested {
			if _, ok := deps[d.Name]; !ok {
				deps[d.Name] = d
			}
		}
	}

	return deps, nil
}

func download(d spec.Dependency, vendorDir string) (*spec.Dependency, error) {
	var p Interface
	switch {
	case d.Source.GitSource != nil:
		p = NewGitPackage(d.Source.GitSource)
	case d.Source.LocalSource != nil:
		p = NewLocalPackage(d.Source.LocalSource)
	}

	if p == nil {
		return nil, errors.New("either git or local source is required")
	}

	version, err := p.Install(context.TODO(), d.Name, vendorDir, d.Version)
	if err != nil {
		return nil, err
	}

	sum := hashDir(filepath.Join(vendorDir, d.Name))

	return &spec.Dependency{
		Name:    d.Name,
		Source:  d.Source,
		Version: version,
		Sum:     sum,
	}, nil
}

func check(d spec.Dependency, vendorDir string) bool {
	if d.Sum == "" {
		// no sum available, need to download
		return false
	}

	dir := filepath.Join(vendorDir, d.Name)
	sum := hashDir(dir)
	return d.Sum == sum
}

func hashDir(dir string) string {
	hasher := sha256.New()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(hasher, f); err != nil {
			return err
		}

		return nil
	})

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
