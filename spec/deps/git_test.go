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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGit(t *testing.T) {
	sshWant := &Dependency{
		Version: "v1",
		Source: Source{GitSource: &Git{
			Scheme: GitSchemeSSH,
			Host:   "github.com",
			User:   "user",
			Repo:   "repo",
			Subdir: "/foobar",
		}},
	}

	tests := []struct {
		name       string
		uri        string
		want       *Dependency
		wantRemote string
	}{
		{
			name: "github-slug",
			uri:  "github.com/ksonnet/ksonnet-lib/ksonnet.beta.3",
			want: &Dependency{
				Version: "master",
				Source: Source{GitSource: &Git{
					Scheme: GitSchemeHTTPS,
					Host:   "github.com",
					User:   "ksonnet",
					Repo:   "ksonnet-lib",
					Subdir: "/ksonnet.beta.3",
				}},
			},
			wantRemote: "https://github.com/ksonnet/ksonnet-lib",
		},
		{
			name:       "ssh.ssh",
			uri:        "ssh://git@github.com/user/repo.git/foobar@v1",
			want:       sshWant,
			wantRemote: "ssh://git@github.com/user/repo.git",
		},
		{
			name:       "ssh.scp",
			uri:        "git@github.com:user/repo.git/foobar@v1",
			want:       sshWant,
			wantRemote: "ssh://git@github.com/user/repo.git", // want ssh format here
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			got := Parse("", c.uri)
			require.NotNilf(t, got, "parsed dependency is nil. Most likely, no regex matched the format.")

			assert.Equal(t, c.want, got)

			require.NotNil(t, got.Source)
			require.NotNil(t, got.Source.GitSource)
			assert.Equal(t, c.wantRemote, got.Source.GitSource.Remote())
		})
	}
}
