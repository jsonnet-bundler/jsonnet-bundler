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
	"github.com/stretchr/testify/require"
)

func TestParseGit(t *testing.T) {
	sshWant := func(host string) *Dependency {
		return &Dependency{
			Version: "v1",
			Source: Source{GitSource: &Git{
				Scheme: GitSchemeSSH,
				Host:   host,
				User:   "user",
				Repo:   "repo",
				Subdir: "/foobar",
			}},
		}
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
			wantRemote: "https://github.com/ksonnet/ksonnet-lib.git",
		},
		{
			name:       "ssh.ssh",
			uri:        "ssh://git@example.com/user/repo.git/foobar@v1",
			want:       sshWant("example.com"),
			wantRemote: "ssh://git@example.com/user/repo.git",
		},
		{
			name:       "ssh.scp",
			uri:        "git@my.host:user/repo.git/foobar@v1",
			want:       sshWant("my.host"),
			wantRemote: "ssh://git@my.host/user/repo.git", // want ssh format here
		},
		{
			name: "ValidGitHTTPS",
			uri:  "https://example.com/foo/bar",
			want: &Dependency{
				Version: "master",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "example.com",
						User:   "foo",
						Repo:   "bar",
						Subdir: "",
					},
				},
			},
			wantRemote: "https://example.com/foo/bar.git",
		},
		{
			name: "ValidGitLabUserGroupHTTPS",
			uri:  "https://gitlab.example.com/first.last/project",
			want: &Dependency{
				Version: "master",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "gitlab.example.com",
						User:   "first.last",
						Repo:   "project",
						Subdir: "",
					},
				},
			},
			wantRemote: "https://gitlab.example.com/first.last/project.git",
		},
		{
			name: "ValidGitNoScheme",
			uri:  "example.com/foo/bar",
			want: &Dependency{
				Version: "master",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "example.com",
						User:   "foo",
						Repo:   "bar",
						Subdir: "",
					},
				},
			},
			wantRemote: "https://example.com/foo/bar.git",
		},
		{
			name: "ValidGitPath",
			uri:  "example.com/foo/bar/baz/bat",
			want: &Dependency{
				Version: "master",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "example.com",
						User:   "foo",
						Repo:   "bar",
						Subdir: "/baz/bat",
					},
				},
			},
			wantRemote: "https://example.com/foo/bar.git",
		},
		{
			name: "ValidGitVersion",
			uri:  "example.com/foo/bar@baz",
			want: &Dependency{
				Version: "baz",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "example.com",
						User:   "foo",
						Repo:   "bar",
						Subdir: "",
					},
				},
			},
			wantRemote: "https://example.com/foo/bar.git",
		},
		{
			name: "ValidGitPathVersion",
			uri:  "example.com/foo/bar/baz@bat",
			want: &Dependency{
				Version: "bat",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "example.com",
						User:   "foo",
						Repo:   "bar",
						Subdir: "/baz",
					},
				},
			},
			wantRemote: "https://example.com/foo/bar.git",
		},
		{
			name: "ValidGitSubdomain",
			uri:  "git.example.com/foo/bar",
			want: &Dependency{
				Version: "master",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "git.example.com",
						User:   "foo",
						Repo:   "bar",
						Subdir: "",
					},
				},
			},
			wantRemote: "https://git.example.com/foo/bar.git",
		},
		{
			name: "ValidGitSubgroups",
			uri:  "example.com/group/subgroup/repository.git",
			want: &Dependency{
				Version: "master",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "example.com",
						User:   "group/subgroup",
						Repo:   "repository",
						Subdir: "",
					},
				},
			},
			wantRemote: "https://example.com/group/subgroup/repository.git",
		},
		{
			name: "ValidGitSubgroupSubDir",
			uri:  "example.com/group/subgroup/repository.git/subdir",
			want: &Dependency{
				Version: "master",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "example.com",
						User:   "group/subgroup",
						Repo:   "repository",
						Subdir: "/subdir",
					},
				},
			},
			wantRemote: "https://example.com/group/subgroup/repository.git",
		},
		{
			name: "ValidBitbucketPersonalRepository",
			uri:  "bitbucket.org/~user/repository.git",
			want: &Dependency{
				Version: "master",
				Source: Source{
					GitSource: &Git{
						Scheme: GitSchemeHTTPS,
						Host:   "bitbucket.org",
						User:   "~user",
						Repo:   "repository",
						Subdir: "",
					},
				},
			},
			wantRemote: "https://bitbucket.org/~user/repository.git",
		},
		{
			name: "InvalidURLInLocalPath",
			uri:  "/home/user/example.com/foo/bar",
			want: nil,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			got := Parse("", c.uri)

			assert.Equal(t, c.want, got)

			if got != nil {
				require.NotNil(t, got.Source)
				require.NotNil(t, got.Source.GitSource)
				assert.Equal(t, c.wantRemote, got.Source.GitSource.Remote())
			}
		})
	}
}
