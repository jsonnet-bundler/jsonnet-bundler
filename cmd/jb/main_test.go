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
	"os"
	"testing"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v3/deps"
	"github.com/stretchr/testify/assert"
)

func TestParseDependency(t *testing.T) {
	const testFolder = "test/jsonnet/foobar"
	err := os.MkdirAll(testFolder, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("test")

	tests := []struct {
		name string
		path string
		want *deps.Dependency
	}{
		{
			name: "Empty",
			path: "",
			want: nil,
		},
		{
			name: "Invalid",
			path: "example.com/foo",
			want: nil,
		},
		{
			name: "GitHTTPS",
			path: "example.com/jsonnet-bundler/jsonnet-bundler",
			want: &deps.Dependency{
				Source: deps.Source{
					GitSource: &deps.Git{
						Scheme: deps.GitSchemeHTTPS,
						Host:   "example.com",
						User:   "jsonnet-bundler",
						Repo:   "jsonnet-bundler",
						Subdir: "",
					},
				},
				Version: "master",
			},
		},
		{
			name: "SSH",
			path: "git+ssh://git@github.com/jsonnet-bundler/jsonnet-bundler.git",
			want: &deps.Dependency{
				Source: deps.Source{
					GitSource: &deps.Git{
						Scheme: deps.GitSchemeSSH,
						Host:   "github.com",
						User:   "jsonnet-bundler",
						Repo:   "jsonnet-bundler",
						Subdir: "",
					},
				},
				Version: "master",
			},
		},
		{
			name: "local",
			path: testFolder,
			want: &deps.Dependency{
				Source: deps.Source{
					LocalSource: &deps.Local{
						Directory: "test/jsonnet/foobar",
					},
				},
				Version: "",
			},
		},
	}
	for _, tt := range tests {
		_ = t.Run(tt.name, func(t *testing.T) {
			dependency := deps.Parse("", tt.path)

			if tt.path == "" {
				assert.Nil(t, dependency)
			} else {
				assert.Equal(t, tt.want, dependency)
			}
		})
	}
}
