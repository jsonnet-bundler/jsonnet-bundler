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

// +build integration

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

func TestInstallCommand(t *testing.T) {
	testcases := []struct {
		Name                    string
		URIs                    []string
		ExpectedCode            int
		ExpectedJsonnetFile     []byte
		ExpectedJsonnetLockFile []byte
	}{
		{
			Name:                "NoURLs",
			ExpectedCode:        0,
			ExpectedJsonnetFile: []byte(`{}`),
		}, {
			Name:                    "OneURL",
			URIs:                    []string{"github.com/jsonnet-bundler/jsonnet-bundler@v0.1.0"},
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"dependencies": [{"name": "jsonnet-bundler", "source": {"git": {"remote": "https://github.com/jsonnet-bundler/jsonnet-bundler", "subdir": ""}}, "version": "v0.1.0"}]}`),
			ExpectedJsonnetLockFile: []byte(`{"dependencies": [{"name": "jsonnet-bundler", "source": {"git": {"remote": "https://github.com/jsonnet-bundler/jsonnet-bundler", "subdir": ""}}, "version": "080f157c7fb85ad0281ea78f6c641eaa570a582f", "sum": "W1uI550rQ66axRpPXA2EZDquyPg/5PHZlvUz1NEzefg="}]}`),
		}, {
			Name:                    "Relative",
			URIs:                    []string{"jsonnet/foobar"},
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"dependencies": [{"name": "foobar", "source": {"local": {"directory": "jsonnet/foobar"}}, "version": ""}]}`),
			ExpectedJsonnetLockFile: []byte(`{"dependencies": [{"name": "foobar", "source": {"local": {"directory": "jsonnet/foobar"}}, "version": ""}]}`),
		},
	}

	localDependency := "jsonnet/foobar"

	cleanup := func() {
		_ = os.Remove(jsonnetfile.File)
		_ = os.Remove(jsonnetfile.LockFile)
		_ = os.RemoveAll("vendor")
		_ = os.RemoveAll("jsonnet")
	}

	for _, tc := range testcases {
		_ = t.Run(tc.Name, func(t *testing.T) {
			cleanup()

			err := os.MkdirAll(localDependency, os.ModePerm)
			assert.NoError(t, err)

			initCommand("")

			jsonnetFileContent(t, jsonnetfile.File, []byte(`{}`))

			installCommand("", "vendor", tc.URIs)

			jsonnetFileContent(t, jsonnetfile.File, tc.ExpectedJsonnetFile)
			if tc.ExpectedJsonnetLockFile != nil {
				jsonnetFileContent(t, jsonnetfile.LockFile, tc.ExpectedJsonnetLockFile)
			}
		})
	}

	cleanup()
}

func jsonnetFileContent(t *testing.T, filename string, content []byte) {
	t.Helper()

	bytes, err := ioutil.ReadFile(filename)
	assert.NoError(t, err)
	if eq := assert.JSONEq(t, string(content), string(bytes)); !eq {
		t.Log(string(bytes))
	}
}

func TestWriteChangedJsonnetFile(t *testing.T) {
	testcases := []struct {
		Name             string
		JsonnetFileBytes []byte
		NewJsonnetFile   spec.JsonnetFile
		ExpectWrite      bool
	}{
		{
			Name:             "NoDiffEmpty",
			JsonnetFileBytes: []byte(`{}`),
			NewJsonnetFile:   spec.New(),
			ExpectWrite:      false,
		},
		{
			Name:             "NoDiffNotEmpty",
			JsonnetFileBytes: []byte(`{"dependencies": [{"name": "foobar"}]}`),
			NewJsonnetFile: spec.JsonnetFile{
				Dependencies: map[string]spec.Dependency{
					"foobar": {
						Name: "foobar",
					},
				},
			},
			ExpectWrite: false,
		},
		{
			Name:             "DiffVersion",
			JsonnetFileBytes: []byte(`{"dependencies": [{"name": "foobar", "version": "1.0"}]}`),
			NewJsonnetFile: spec.JsonnetFile{
				Dependencies: map[string]spec.Dependency{
					"foobar": {
						Name:    "foobar",
						Version: "2.0",
					},
				},
			},
			ExpectWrite: true,
		},
		{
			Name:             "Diff",
			JsonnetFileBytes: []byte(`{}`),
			NewJsonnetFile: spec.JsonnetFile{
				Dependencies: map[string]spec.Dependency{
					"foobar": {
						Name: "foobar",
						Source: spec.Source{
							GitSource: &spec.GitSource{
								Remote: "https://github.com/foobar/foobar",
								Subdir: "",
							},
						},
						Version:   "master",
						DepSource: "",
					}},
			},
			ExpectWrite: true,
		},
	}
	outputjsonnetfile := "changedjsonnet.json"
	for _, tc := range testcases {
		_ = t.Run(tc.Name, func(t *testing.T) {
			clean := func() {
				_ = os.Remove(outputjsonnetfile)
			}
			clean()
			defer clean()

			err := writeChangedJsonnetFile(tc.JsonnetFileBytes, &tc.NewJsonnetFile, outputjsonnetfile)
			assert.NoError(t, err)

			if tc.ExpectWrite {
				assert.FileExists(t, outputjsonnetfile)
			} else {
				_, err := os.Lstat(outputjsonnetfile)
				if err != nil {
					assert.True(t, os.IsNotExist(err))
				}
			}
		})
	}
}
