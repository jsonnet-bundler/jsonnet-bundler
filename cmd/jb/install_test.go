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
	"path/filepath"
	"testing"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/stretchr/testify/assert"
)

func TestInstallCommand(t *testing.T) {
	testcases := []struct {
		Name                    string
		URLs                    []string
		ExpectedCode            int
		ExpectedJsonnetFile     []byte
		ExpectedJsonnetLockFile []byte
	}{
		{
			Name:                    "NoURLs",
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"dependencies":null}`),
			ExpectedJsonnetLockFile: []byte(`{"dependencies":null}`),
		}, {
			Name:                    "OneURL",
			URLs:                    []string{"github.com/jsonnet-bundler/jsonnet-bundler@v0.1.0"},
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"dependencies": [{"name": "jsonnet-bundler", "source": {"git": {"remote": "https://github.com/jsonnet-bundler/jsonnet-bundler", "subdir": ""}}, "version": "v0.1.0"}]}`),
			ExpectedJsonnetLockFile: []byte(`{"dependencies": [{"name": "jsonnet-bundler", "source": {"git": {"remote": "https://github.com/jsonnet-bundler/jsonnet-bundler", "subdir": ""}}, "version": "080f157c7fb85ad0281ea78f6c641eaa570a582f"}]}`),
		}, {
			Name:                    "Relative",
			URLs:                    []string{"test/jsonnet/foobar"},
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"dependencies": [{"name": "foobar", "source": {"local": {"directory": "test/jsonnet/foobar"}}, "version": ""}]}`),
			ExpectedJsonnetLockFile: []byte(`{"dependencies": [{"name": "foobar", "source": {"local": {"directory": "test/jsonnet/foobar"}}, "version": ""}]}`),
		},
	}

	for _, tc := range testcases {
		_ = t.Run(tc.Name, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", "jb-install")
			assert.NoError(t, err)
			err = os.MkdirAll(filepath.Join(tempDir, "test/jsonnet/foobar"), os.ModePerm)
			assert.NoError(t, err)
			defer os.Remove(tempDir)
			defer os.RemoveAll("vendor") // cloning jsonnet-bundler will create this folder

			jsonnetFile := filepath.Join(tempDir, jsonnetfile.File)
			jsonnetLockFile := filepath.Join(tempDir, jsonnetfile.LockFile)

			code := initCommand(tempDir)
			assert.Equal(t, 0, code)

			jsonnetFileContent(t, jsonnetFile, []byte(`{}`))

			code = installCommand(tempDir, "vendor", tc.URLs...)
			assert.Equal(t, tc.ExpectedCode, code)

			jsonnetFileContent(t, jsonnetFile, tc.ExpectedJsonnetFile)
			jsonnetFileContent(t, jsonnetLockFile, tc.ExpectedJsonnetLockFile)
		})
	}
}

func jsonnetFileContent(t *testing.T, filename string, content []byte) {
	t.Helper()

	bytes, err := ioutil.ReadFile(filename)
	assert.NoError(t, err)
	if eq := assert.JSONEq(t, string(content), string(bytes)); !eq {
		t.Log(string(bytes))
	}
}
