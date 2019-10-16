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

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg"
	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

func installCommand(dir, jsonnetHome string, uris []string) int {
	if dir == "" {
		dir = "."
	}

	kingpin.FatalIfError(
		os.MkdirAll(filepath.Join(dir, jsonnetHome, ".tmp"), os.ModePerm),
		"creating vendor folder")

	jsonnetFile, err := jsonnetfile.Load(filepath.Join(dir, jsonnetfile.File))
	kingpin.FatalIfError(err, "failed to load jsonnetfile")

	for _, u := range uris {
		d := parseDependency(dir, u)
		jsonnetFile.Dependencies[d.Name] = *d
	}

	lockFile, err := jsonnetfile.Load(filepath.Join(dir, jsonnetfile.LockFile))
	if !os.IsNotExist(err) {
		kingpin.FatalIfError(err, "failed to load lockfile")
	}

	locked, err := pkg.Ensure(jsonnetFile, jsonnetHome, lockFile.Dependencies)
	kingpin.FatalIfError(err, "failed to install packages")

	kingpin.FatalIfError(
		writeJSONFile(filepath.Join(dir, jsonnetfile.File), jsonnetFile),
		"updating jsonnetfile.json")
	kingpin.FatalIfError(
		writeJSONFile(filepath.Join(dir, jsonnetfile.LockFile), spec.JsonnetFile{Dependencies: locked}),
		"updating jsonnetfile.lock.json")

	return 0
}

func writeJSONFile(name string, d interface{}) error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return errors.Wrap(err, "encoding json")
	}
	b = append(b, []byte("\n")...)

	return ioutil.WriteFile(name, b, 0644)
}
