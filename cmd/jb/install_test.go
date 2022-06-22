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

//go:build integration
// +build integration

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	v1 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v1"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

const initContents = `{"version": 1, "dependencies": [], "legacyImports": true}`

func TestInstallCommand(t *testing.T) {
	testInstallCommandWithJsonnetHome(t, "vendor")
}

func TestInstallCommandCustomJsonnetHome(t *testing.T) {
	testInstallCommandWithJsonnetHome(t, "custom-vendor-dir")
}

func TestInstallCommandDeepCustomJsonnetHome(t *testing.T) {
	testInstallCommandWithJsonnetHome(t, "custom/vendor/dir")
}

func testInstallCommandWithJsonnetHome(t *testing.T, jsonnetHome string) {
	testcases := []struct {
		Name                    string
		URIs                    []string
		ExpectedCode            int
		ExpectedJsonnetFile     []byte
		ExpectedJsonnetLockFile []byte
		single                  bool
	}{
		{
			Name:                "NoURLs",
			ExpectedCode:        0,
			ExpectedJsonnetFile: []byte(initContents),
		},
		{
			Name:                    "OneURL",
			URIs:                    []string{"github.com/jsonnet-bundler/jsonnet-bundler@v0.1.0"},
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"version": 1, "dependencies": [{"source": {"git": {"remote": "https://github.com/jsonnet-bundler/jsonnet-bundler.git", "subdir": ""}}, "version": "v0.1.0"}], "legacyImports": true}`),
			ExpectedJsonnetLockFile: []byte(`{"version": 1, "dependencies": [{"source": {"git": {"remote": "https://github.com/jsonnet-bundler/jsonnet-bundler.git", "subdir": ""}}, "version": "080f157c7fb85ad0281ea78f6c641eaa570a582f", "sum": "W1uI550rQ66axRpPXA2EZDquyPg/5PHZlvUz1NEzefg="}], "legacyImports": false}`),
		},
		{
			Name:                    "Local",
			URIs:                    []string{"jsonnet/foobar"},
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"version": 1, "dependencies": [{"source": {"local": {"directory": "jsonnet/foobar"}}, "version": ""}], "legacyImports": true}`),
			ExpectedJsonnetLockFile: []byte(`{"version": 1, "dependencies": [{"source": {"local": {"directory": "jsonnet/foobar"}}, "version": ""}], "legacyImports": false}`),
		},
		{
			Name:                    "single",
			URIs:                    []string{"github.com/grafana/loki/production/ksonnet/loki@bd4d516262c107a0bde7a962fa2b1e567a2c21e5"},
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/grafana/loki.git","subdir":"production/ksonnet/loki"}},"version":"bd4d516262c107a0bde7a962fa2b1e567a2c21e5","single":true}],"legacyImports":true}`),
			ExpectedJsonnetLockFile: []byte(`{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/grafana/loki.git","subdir":"production/ksonnet/loki"}},"version":"bd4d516262c107a0bde7a962fa2b1e567a2c21e5","sum":"ExovUKXmZ4KwJAv/q8ZwNW9BdIZlrxmoGrne7aR64wo=","single":true}],"legacyImports":false}`),
			single:                  true,
		},
	}

	localDependency := "jsonnet/foobar"

	cleanup := func() {
		_ = os.Remove(jsonnetfile.File)
		_ = os.Remove(jsonnetfile.LockFile)
		_ = os.RemoveAll(jsonnetHome)
		_ = os.RemoveAll("jsonnet")
	}

	for _, tc := range testcases {
		_ = t.Run(tc.Name, func(t *testing.T) {
			cleanup()

			err := os.MkdirAll(localDependency, os.ModePerm)
			assert.NoError(t, err)

			// init + check it works correctly (legacyImports true, empty dependencies)
			initCommand("")
			jsonnetFileContent(t, jsonnetfile.File, []byte(initContents))

			// install something, check it writes only if required, etc.
			installCommand("", jsonnetHome, tc.URIs, tc.single, "")
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
		NewJsonnetFile   v1.JsonnetFile
		ExpectWrite      bool
	}{
		{
			Name:             "NoDiffEmpty",
			JsonnetFileBytes: []byte(`{}`),
			NewJsonnetFile:   v1.New(),
			ExpectWrite:      false,
		},
		{
			Name:             "NoDiffNotEmpty",
			JsonnetFileBytes: []byte(`{"dependencies": [{"version": "master"}]}`),
			NewJsonnetFile: v1.JsonnetFile{
				Dependencies: map[string]deps.Dependency{
					"": {
						Version: "master",
					},
				},
			},
			ExpectWrite: false,
		},
		{
			Name:             "DiffVersion",
			JsonnetFileBytes: []byte(`{"dependencies": [{"version": "1.0"}]}`),
			NewJsonnetFile: v1.JsonnetFile{
				Dependencies: map[string]deps.Dependency{
					"": {
						Version: "2.0",
					},
				},
			},
			ExpectWrite: true,
		},
		{
			Name:             "Diff",
			JsonnetFileBytes: []byte(`{}`),
			NewJsonnetFile: v1.JsonnetFile{
				Dependencies: map[string]deps.Dependency{
					"github.com/foobar/foobar": {
						Source: deps.Source{
							GitSource: &deps.Git{
								Scheme: deps.GitSchemeHTTPS,
								Host:   "github.com",
								User:   "foobar",
								Repo:   "foobar",
								Subdir: "",
							},
						},
						Version: "master",
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
