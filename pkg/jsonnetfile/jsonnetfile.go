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

package jsonnetfile

import (
	"encoding/json"
	"io/ioutil"
	"os"

	v2 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v2"
	v3 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v3"
	depsv3 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v3/deps"
	"github.com/pkg/errors"
)

const (
	File     = "jsonnetfile.json"
	LockFile = "jsonnetfile.lock.json"
)

var (
	ErrNoFile   = errors.New("no jsonnetfile")
	ErrUpdateJB = errors.New("jsonnetfile version unknown, update jb")
)

// Load reads a jsonnetfile.(lock).json from disk
func Load(filepath string) (v3.JsonnetFile, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return v3.New(), err
	}

	return Unmarshal(bytes)
}

// Unmarshal creates a spec.JsonnetFile from bytes. Empty bytes
// will create an empty spec.
func Unmarshal(bytes []byte) (v3.JsonnetFile, error) {
	m := v3.New()

	if len(bytes) == 0 {
		return m, nil
	}

	versions := struct {
		Version float64 `json:"version"`
	}{}

	err := json.Unmarshal(bytes, &versions)
	if err != nil {
		return m, err
	}

	if versions.Version > 3 {
		return m, ErrUpdateJB
	}

	if versions.Version == 3 {
		if err := json.Unmarshal(bytes, &m); err != nil {
			return m, errors.Wrap(err, "failed to unmarshal v3 file")
		}

		return m, nil
	} else {
		var mv2 v2.JsonnetFile
		if err := json.Unmarshal(bytes, &mv2); err != nil {
			return m, errors.Wrap(err, "failed to unmarshal v2 file")
		}

		for name, dep := range mv2.Dependencies {
			var d depsv3.Dependency
			if dep.Source.GitSource != nil {
				d = *depsv3.Parse("", dep.Source.GitSource.Remote)
				d.Source.GitSource.Subdir = dep.Source.GitSource.Subdir
			}
			if dep.Source.LocalSource != nil {
				d = *depsv3.Parse(dep.Source.LocalSource.Directory, dep.Source.GitSource.Remote)
			}

			d.Sum = dep.Sum
			d.Version = dep.Version

			m.Dependencies[name] = d
		}

		return m, nil
	}
}

// Exists returns whether the file at the given path exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
