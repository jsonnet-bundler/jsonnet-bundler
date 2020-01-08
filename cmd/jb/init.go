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
	"path/filepath"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

func initCommand(dir string) int {
	exists, err := jsonnetfile.Exists(jsonnetfile.File)
	kingpin.FatalIfError(err, "Failed to check for jsonnetfile.json")

	if exists {
		kingpin.Errorf("jsonnetfile.json already exists")
		return 1
	}

	// default to go-style only for new setups
	s := spec.New()
	s.LegacyImports = false

	contents, err := json.MarshalIndent(s, "", "  ")
	kingpin.FatalIfError(err, "formatting jsonnetfile contents as json")
	contents = append(contents, []byte("\n")...)

	filename := filepath.Join(dir, jsonnetfile.File)

	ioutil.WriteFile(filename, contents, 0644)
	kingpin.FatalIfError(err, "Failed to write new jsonnetfile.json")

	return 0
}
