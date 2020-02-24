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
package deps

import (
	"os"
	"testing"

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
		want *Dependency
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
			name: "InvalidDomain",
			path: "example.c/foo/bar",
			want: nil,
		},
		{
			name: "InvalidDomain2",
			path: "examplec/foo/bar",
			want: nil,
		},
		{
			name: "local",
			path: testFolder,
			want: &Dependency{
				Source: Source{
					LocalSource: &Local{
						Directory: "test/jsonnet/foobar",
					},
				},
				Version: "",
			},
		},
	}
	for _, tt := range tests {
		_ = t.Run(tt.name, func(t *testing.T) {
			dependency := Parse("", tt.path)

			if tt.path == "" {
				assert.Nil(t, dependency)
			} else {
				assert.Equal(t, tt.want, dependency)
			}
		})
	}
}
