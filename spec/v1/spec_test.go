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

package spec

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

const jsonJF = `{
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

func testData() JsonnetFile {
	return JsonnetFile{
		LegacyImports: false,
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
	}
}

// TestUnmarshal checks that unmarshalling works
func TestUnmarshal(t *testing.T) {
	var dst JsonnetFile
	err := json.Unmarshal([]byte(jsonJF), &dst)
	require.NoError(t, err)
	assert.Equal(t, testData(), dst)
}

// TestMarshal checks that marshalling works
func TestMarshal(t *testing.T) {
	data, err := json.Marshal(testData())
	require.NoError(t, err)
	assert.JSONEq(t, jsonJF, string(data))
}

// TestRemarshal checks that unmarshalling a previously marshalled object yields
// the same object
func TestRemarshal(t *testing.T) {
	jf := testData()

	data, err := json.Marshal(jf)
	require.NoError(t, err)

	var dst JsonnetFile
	err = json.Unmarshal(data, &dst)
	require.NoError(t, err)

	assert.Equal(t, jf, dst)
}
