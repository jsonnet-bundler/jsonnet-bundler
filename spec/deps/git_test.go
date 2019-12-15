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
