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

	errors "github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

var (
	FailedToCheck = errors.New("failed to check for jsonnet file:")
	AlreadyExists = errors.New("jsonnet file already exists")
	FailedToWrite = errors.New("failed to write jsonnet file:")
)

func initCommand(dir string) int {
	err := initOperation(dir)
	if err != nil {
		kingpin.Errorf("Failed to initialize: %v", err)
		return 1
	}
	return 0
}

func initOperation(dir string) error {
	exists, err := jsonnetfile.Exists(jsonnetfile.File)
	if err != nil {
		return errors.Wrap(FailedToCheck, err.Error())
	}

	if exists {
		return AlreadyExists
	}

	s := spec.New()
	// TODO: disable them by default eventually
	// s.LegacyImports = false

	contents, err := json.MarshalIndent(s, "", "  ")
	kingpin.FatalIfError(err, "formatting jsonnetfile contents as json")
	contents = append(contents, []byte("\n")...)

	filename := filepath.Join(dir, jsonnetfile.File)

	if err := ioutil.WriteFile(filename, contents, 0644); err != nil {
		return errors.Wrap(FailedToWrite, err.Error())
	}

	return nil
}
