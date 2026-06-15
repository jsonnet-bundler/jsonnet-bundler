// Copyright 2018 jsonnet-bundler authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package deps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOCI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want *Dependency
	}{
		{
			name: "DefaultTag",
			uri:  "oci://registry.example.com/namespace/name",
			want: &Dependency{
				Version: "latest",
				Source: Source{
					OCISource: &OCI{Repo: "registry.example.com/namespace/name"},
				},
			},
		},
		{
			name: "ExplicitTag",
			uri:  "oci://registry.example.com/namespace/name:v1.2.3",
			want: &Dependency{
				Version: "v1.2.3",
				Source: Source{
					OCISource: &OCI{Repo: "registry.example.com/namespace/name"},
				},
			},
		},
		{
			name: "Digest",
			uri:  "oci://registry.example.com/namespace/name@sha256:0000000000000000000000000000000000000000000000000000000000000000",
			want: &Dependency{
				Version: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
				Source: Source{
					OCISource: &OCI{Repo: "registry.example.com/namespace/name"},
				},
			},
		},
		{
			name: "RegistryWithPort",
			uri:  "oci://localhost:5000/name:dev",
			want: &Dependency{
				Version: "dev",
				Source: Source{
					OCISource: &OCI{Repo: "localhost:5000/name"},
				},
			},
		},
		{
			name: "Invalid",
			uri:  "oci://",
			want: nil,
		},
	}

	for _, tt := range tests {
		_ = t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Parse("", tt.uri))
		})
	}

	// A non-oci:// URI must not be claimed by the OCI parser; Parse falls
	// through to git/local for those.
	assert.Nil(t, parseOCI("github.com/foo/bar"))
}

func TestOCINames(t *testing.T) {
	o := &OCI{Repo: "registry.example.com/namespace/name"}
	assert.Equal(t, "registry.example.com/namespace/name", o.Name())
	assert.Equal(t, "name", o.LegacyName())
}
