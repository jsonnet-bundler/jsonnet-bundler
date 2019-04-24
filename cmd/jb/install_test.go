// +build integration

package main

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/stretchr/testify/assert"
)

func TestInstallCommand(t *testing.T) {
	testcases := []struct {
		Name                    string
		URLs                    []*url.URL
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
			Name: "OneURL",
			URLs: []*url.URL{
				{
					Scheme: "https",
					Host:   "github.com",
					Path:   "jsonnet-bundler/jsonnet-bundler",
				},
			},
			ExpectedCode:            0,
			ExpectedJsonnetFile:     []byte(`{"dependencies": [{"name": "jsonnet-bundler", "source": {"git": {"remote": "https://github.com/jsonnet-bundler/jsonnet-bundler", "subdir": ""}}, "version": "master"}]}`),
			ExpectedJsonnetLockFile: []byte(`{"dependencies": [{"name": "jsonnet-bundler", "source": {"git": {"remote": "https://github.com/jsonnet-bundler/jsonnet-bundler", "subdir": ""}}, "version": "080f157c7fb85ad0281ea78f6c641eaa570a582f"}]}`),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", "jb-install")
			assert.NoError(t, err)
			defer os.Remove(tempDir)

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
	bytes, err := ioutil.ReadFile(filename)
	assert.NoError(t, err)
	if eq := assert.JSONEq(t, string(content), string(bytes)); !eq {
		t.Log(string(bytes))
	}
}
