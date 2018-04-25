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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/pkg/errors"
)

var (
	JsonnetFile     = "jsonnetfile.json"
	JsonnetLockFile = "jsonnetfile.lock.json"
	VersionMismatch = errors.New("multiple colliding versions specified")
)

func Install(ctx context.Context, m spec.JsonnetFile, dir string) (lock *spec.JsonnetFile, err error) {
	lock = &spec.JsonnetFile{}
	for _, dep := range m.Dependencies {
		tmp := filepath.Join(dir, ".tmp")
		err = os.MkdirAll(tmp, os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create general tmp dir")
		}
		tmpDir, err := ioutil.TempDir(tmp, fmt.Sprintf("jsonnetpkg-%s-%s", dep.Name, dep.Version))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tmp dir")
		}
		defer os.RemoveAll(tmpDir)

		subdir := ""
		var p Interface
		if dep.Source.GitSource != nil {
			p = NewGitPackage(dep.Source.GitSource)
			subdir = dep.Source.GitSource.Subdir
		}

		lockVersion, err := p.Install(ctx, tmpDir, dep.Version)
		if err != nil {
			return nil, errors.Wrap(err, "failed to install package")
		}
		// need to deduplicate/error when multiple entries
		lock.Dependencies, err = insertDependency(lock.Dependencies, spec.Dependency{
			Name:    dep.Name,
			Source:  dep.Source,
			Version: lockVersion,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert dependency to lock dependencies")
		}

		destPath := path.Join(dir, dep.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to find destination path for package")
		}

		err = os.MkdirAll(path.Dir(destPath), os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create parent path")
		}

		err = os.RemoveAll(destPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to clean previous destination path")
		}
		err = os.Rename(path.Join(tmpDir, subdir), destPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to move package")
		}

		if _, err := os.Stat(path.Join(destPath, JsonnetFile)); !os.IsNotExist(err) {
			depsDeps, err := LoadJsonnetfile(path.Join(destPath, JsonnetFile))
			if err != nil {
				return nil, err
			}

			depsInstalledByDependency, err := Install(ctx, depsDeps, dir)
			if err != nil {
				return nil, err
			}

			for _, d := range depsInstalledByDependency.Dependencies {
				lock.Dependencies, err = insertDependency(lock.Dependencies, d)
				if err != nil {
					return nil, errors.Wrap(err, "failed to insert dependency to lock dependencies")
				}
			}
		}
	}

	return lock, nil
}

func insertDependency(deps []spec.Dependency, newDep spec.Dependency) ([]spec.Dependency, error) {
	res := []spec.Dependency{}
	for _, d := range deps {
		if d.Name == newDep.Name {
			if d.Version != newDep.Version {
				return nil, VersionMismatch
			}
			res = append(res, d)
		} else {
			res = append(res, d)
		}
	}

	return res, nil
}

func LoadJsonnetfile(filename string) (spec.JsonnetFile, error) {
	m := spec.JsonnetFile{}

	f, err := os.Open(filename)
	if err != nil {
		return m, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&m)
	if err != nil {
		return m, err
	}

	return m, nil
}
