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

package jsonnetfile_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	v1 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v1"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

const notExist = "/this/does/not/exist"

const v0JSON = `{
  "dependencies": [
    {
      "name": "grafana-builder",
      "source": {
        "git": {
          "remote": "https://github.com/grafana/jsonnet-libs",
          "subdir": "grafana-builder"
        }
      },
      "version": "54865853ebc1f901964e25a2e7a0e4d2cb6b9648",
      "sum": "ELsYwK+kGdzX1mee2Yy+/b2mdO4Y503BOCDkFzwmGbE="
    },
    {
      "name": "prometheus-mixin",
      "source": {
        "git": {
          "remote": "https://github.com/prometheus/prometheus",
          "subdir": "documentation/prometheus-mixin"
        }
      },
      "version": "7c039a6b3b4b2a9d7c613ac8bd3fc16e8ca79684",
      "sum": "bVGOsq3hLOw2irNPAS91a5dZJqQlBUNWy3pVwM4+kIY="
    }
  ]
}`

var v0Jsonnetfile = v1.JsonnetFile{
	Dependencies: map[string]deps.Dependency{
		"grafana-builder": {
			Source: deps.Source{
				GitSource: &deps.Git{
					Scheme: deps.GitSchemeHTTPS,
					Host:   "github.com",
					User:   "grafana",
					Repo:   "jsonnet-libs",
					Subdir: "grafana-builder",
				},
			},
			Version: "54865853ebc1f901964e25a2e7a0e4d2cb6b9648",
			Sum:     "ELsYwK+kGdzX1mee2Yy+/b2mdO4Y503BOCDkFzwmGbE=",
		},
		"prometheus-mixin": {
			Source: deps.Source{
				GitSource: &deps.Git{
					Scheme: deps.GitSchemeHTTPS,
					Host:   "github.com",
					User:   "prometheus",
					Repo:   "prometheus",
					Subdir: "documentation/prometheus-mixin",
				},
			},
			Version: "7c039a6b3b4b2a9d7c613ac8bd3fc16e8ca79684",
			Sum:     "bVGOsq3hLOw2irNPAS91a5dZJqQlBUNWy3pVwM4+kIY=",
		},
	},
	LegacyImports: true,
}

const v1JSON = `{
  "version": 1,
  "dependencies": [
    {
      "source": {
        "git": {
          "remote": "https://github.com/grafana/jsonnet-libs",
          "subdir": "grafana-builder"
        }
      },
      "version": "54865853ebc1f901964e25a2e7a0e4d2cb6b9648",
      "sum": "ELsYwK+kGdzX1mee2Yy+/b2mdO4Y503BOCDkFzwmGbE="
    },
    {
      "name": "prometheus",
      "source": {
        "git": {
          "remote": "https://github.com/prometheus/prometheus",
          "subdir": "documentation/prometheus-mixin"
        }
      },
      "version": "7c039a6b3b4b2a9d7c613ac8bd3fc16e8ca79684",
      "sum": "bVGOsq3hLOw2irNPAS91a5dZJqQlBUNWy3pVwM4+kIY="
    }
  ],
  "legacyImports": false
}`

var v1Jsonnetfile = v1.JsonnetFile{
	Dependencies: map[string]deps.Dependency{
		"github.com/grafana/jsonnet-libs/grafana-builder": {
			Source: deps.Source{
				GitSource: &deps.Git{
					Scheme: deps.GitSchemeHTTPS,
					Host:   "github.com",
					User:   "grafana",
					Repo:   "jsonnet-libs",
					Subdir: "/grafana-builder",
				},
			},
			Version: "54865853ebc1f901964e25a2e7a0e4d2cb6b9648",
			Sum:     "ELsYwK+kGdzX1mee2Yy+/b2mdO4Y503BOCDkFzwmGbE=",
		},
		"github.com/prometheus/prometheus/documentation/prometheus-mixin": {
			LegacyNameCompat: "prometheus",
			Source: deps.Source{
				GitSource: &deps.Git{
					Scheme: deps.GitSchemeHTTPS,
					Host:   "github.com",
					User:   "prometheus",
					Repo:   "prometheus",
					Subdir: "/documentation/prometheus-mixin",
				},
			},
			Version: "7c039a6b3b4b2a9d7c613ac8bd3fc16e8ca79684",
			Sum:     "bVGOsq3hLOw2irNPAS91a5dZJqQlBUNWy3pVwM4+kIY=",
		},
	},
	LegacyImports: false,
}

func TestVersions(t *testing.T) {
	tests := []struct {
		Name        string
		JSON        string
		Jsonnetfile v1.JsonnetFile
		Error       error
	}{
		{
			Name:        "v0",
			JSON:        v0JSON,
			Jsonnetfile: v0Jsonnetfile,
		},
		{
			Name:        "v1",
			JSON:        v1JSON,
			Jsonnetfile: v1Jsonnetfile,
		},
		{
			Name:        "v100",
			JSON:        `{"version": 100}`,
			Jsonnetfile: v1.New(),
			Error:       jsonnetfile.ErrUpdateJB,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			jf, err := jsonnetfile.Unmarshal([]byte(tc.JSON))
			assert.Equal(t, tc.Error, err)
			assert.Equal(t, tc.Jsonnetfile, jf)
		})
	}
}

func TestLoadV1(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "jb-load-jsonnetfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, jsonnetfile.File)
	err = ioutil.WriteFile(tempFile, []byte(v1JSON), os.ModePerm)
	assert.Nil(t, err)

	jf, err := jsonnetfile.Load(tempFile)
	assert.Nil(t, err)
	assert.Equal(t, v1Jsonnetfile, jf)
}

func TestLoadEmpty(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "jb-load-empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// write empty json file
	tempFile := filepath.Join(tempDir, jsonnetfile.File)
	err = ioutil.WriteFile(tempFile, []byte(`{"version":1}`), os.ModePerm)
	assert.Nil(t, err)

	// expect it to be loaded properly
	got, err := jsonnetfile.Load(tempFile)
	assert.Nil(t, err)
	assert.Equal(t, v1.New(), got)
}

func TestLoadNotExist(t *testing.T) {
	jf, err := jsonnetfile.Load(notExist)
	assert.Equal(t, v1.New(), jf)
	assert.Error(t, err)
}

func TestFileExists(t *testing.T) {
	{
		exists, err := jsonnetfile.Exists(notExist)
		assert.False(t, exists)
		assert.Nil(t, err)
	}
	{
		tempFile, err := ioutil.TempFile("", "jb-exists")
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			err := os.Remove(tempFile.Name())
			assert.Nil(t, err)
		}()

		exists, err := jsonnetfile.Exists(tempFile.Name())
		assert.True(t, exists)
		assert.Nil(t, err)
	}
}
