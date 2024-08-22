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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
				Dependencies: addDependencies(deps.NewOrdered(),
					deps.Dependency{
						Version: "master",
					}),
			},
			ExpectWrite: false,
		},
		{
			Name:             "DiffVersion",
			JsonnetFileBytes: []byte(`{"dependencies": [{"version": "1.0"}]}`),
			NewJsonnetFile: v1.JsonnetFile{
				Dependencies: addDependencies(deps.NewOrdered(),
					deps.Dependency{
						Version: "2.0",
					}),
			},
			ExpectWrite: true,
		},
		{
			Name:             "Diff",
			JsonnetFileBytes: []byte(`{}`),
			NewJsonnetFile: v1.JsonnetFile{
				Dependencies: addDependencies(deps.NewOrdered(),
					deps.Dependency{
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
					}),
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

func TestInstallTransitive(t *testing.T) {
	const (
		frozenLibFirstCommit  = "9f40207f668e382b706e1822f2d46ce2cd0a57cc"
		frozenLibSecondCommit = "ed7c1aff9e10d3b42fb130446d495f1c769ecd7b"
	)

	baseDir := t.TempDir()
	subDirA := filepath.Join(baseDir, "a")
	subDirB := filepath.Join(baseDir, "b")

	writeDepFileTree(t, map[string]v1.JsonnetFile{
		baseDir: {
			Dependencies: addDependencies(deps.NewOrdered(),
				localDependency(subDirA),
				localDependency(subDirB),
			)},
		subDirA: jsonnetFileWithFrozenLib(frozenLibFirstCommit, ""),
		subDirB: jsonnetFileWithFrozenLib(frozenLibSecondCommit, ""),
	})

	require.Equal(t, 0, installCommand(baseDir, "vendor", nil, false, ""))

	lockCheckFrozenLibVersion(t, filepath.Join(baseDir, "jsonnetfile.lock.json"), frozenLibFirstCommit)
	require.NoError(t, os.RemoveAll(filepath.Join(baseDir, "jsonnetfile.lock.json")))

	// Reverse the order of the dependencies. The lock file should now contain the second commit in its version field.
	writeDepFileTree(t, map[string]v1.JsonnetFile{
		subDirA: jsonnetFileWithFrozenLib(frozenLibSecondCommit, ""),
		subDirB: jsonnetFileWithFrozenLib(frozenLibFirstCommit, ""),
	})

	require.Equal(t, 0, installCommand(baseDir, "vendor", nil, false, ""))

	lockCheckFrozenLibVersion(t, filepath.Join(baseDir, "jsonnetfile.lock.json"), frozenLibSecondCommit)
}

func lockCheckFrozenLibVersion(t *testing.T, lockPath, version string) {
	t.Helper()

	rawLock, err := os.ReadFile(lockPath)
	require.NoError(t, err)
	var lock v1.JsonnetFile
	require.NoError(t, json.Unmarshal([]byte(rawLock), &lock))

	lf, lfExists := lock.Dependencies.Get("github.com/jsonnet-bundler/frozen-lib")
	require.True(t, lfExists, "expected to find frozen-lib in lock file")
	require.Equal(t, version, lf.Version, "lock file: expected frozen-lib to have commit version of the first dependency in the base jsonnet file")
}

func addDependencies(o *deps.Ordered, ds ...deps.Dependency) *deps.Ordered {
	for _, d := range ds {
		o.Set(d.Name(), d)
	}
	return o
}

func jsonnetFileWithFrozenLib(version, sum string) v1.JsonnetFile {
	return v1.JsonnetFile{
		Dependencies: addDependencies(deps.NewOrdered(),
			frozenDependency(version, sum)),
	}
}

func frozenDependency(version, sum string) deps.Dependency {
	return deps.Dependency{
		Source: deps.Source{
			GitSource: &deps.Git{
				Scheme: "https://",
				Host:   "github.com",
				User:   "jsonnet-bundler",
				Repo:   "frozen-lib",
			},
		},
		Version: version,
		Sum:     sum,
	}
}

func localDependency(dir string) deps.Dependency {
	return deps.Dependency{
		Source: deps.Source{
			LocalSource: &deps.Local{
				Directory: dir,
			},
		},
	}
}

func writeDepFileTree(t *testing.T, files map[string]v1.JsonnetFile) {
	t.Helper()

	for dir, file := range files {
		require.NoError(t, os.MkdirAll(dir, os.ModePerm))
		rj, err := json.Marshal(file)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "jsonnetfile.json"), rj, 0644))
	}
}
