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
)

func TestParseGit(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		want       *Dependency
		wantRemote string
	}{
		{
			name: "GitHub",
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
			name: "SSH",
			uri:  "git+ssh://git@my.host:user/repo.git/foobar@v1",
			want: &Dependency{
				Version: "v1",
				Source: Source{GitSource: &Git{
					Scheme: GitSchemeSSH,
					Host:   "my.host",
					User:   "user",
					Repo:   "repo",
					Subdir: "/foobar",
				}},
			},
			wantRemote: "ssh://git@my.host:user/repo.git",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			got := Parse("", c.uri)
			assert.Equal(t, c.want, got)
			assert.Equal(t, c.wantRemote, got.Source.GitSource.Remote())
		})
	}
}
