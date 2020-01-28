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
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	GitSchemeSSH   = "ssh://git@"
	GitSchemeHTTPS = "https://"
)

// Git holds all required information for cloning a package from git
type Git struct {
	// Scheme (Protocol) used (https, git+ssh)
	Scheme string

	// Hostname the repo is located at
	Host string
	// User (github.com/<user>)
	User string
	// Repo (github.com/<user>/<repo>)
	Repo string
	// Subdir (github.com/<user>/<repo>/<subdir>)
	Subdir string
}

// json representation of Git (for compatiblity with old format)
type jsonGit struct {
	Remote string `json:"remote"`
	Subdir string `json:"subdir"`
}

// MarshalJSON takes care of translating between Git and jsonGit
func (gs *Git) MarshalJSON() ([]byte, error) {
	j := jsonGit{
		Remote: gs.Remote(),
		Subdir: strings.TrimPrefix(gs.Subdir, "/"),
	}
	return json.Marshal(j)
}

// UnmarshalJSON takes care of translating between Git and jsonGit
func (gs *Git) UnmarshalJSON(data []byte) error {
	var j jsonGit
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	if j.Subdir != "" {
		gs.Subdir = "/" + strings.TrimPrefix(j.Subdir, "/")
	}

	tmp := parseGit(j.Remote)
	gs.Host = tmp.Source.GitSource.Host
	gs.User = tmp.Source.GitSource.User
	gs.Repo = tmp.Source.GitSource.Repo
	gs.Scheme = tmp.Source.GitSource.Scheme
	return nil
}

// Name returns the repository in a go-like format (github.com/user/repo/subdir)
func (gs *Git) Name() string {
	return fmt.Sprintf("%s/%s/%s%s", gs.Host, gs.User, gs.Repo, gs.Subdir)
}

// LegacyName returns the last element of the packages path
// example: github.com/ksonnet/ksonnet-lib/ksonnet.beta.4 becomes ksonnet.beta.4
func (gs *Git) LegacyName() string {
	return filepath.Base(gs.Repo + gs.Subdir)
}

var gitProtoFmts = map[string]string{
	GitSchemeSSH:   GitSchemeSSH + "%s/%s/%s.git",
	GitSchemeHTTPS: GitSchemeHTTPS + "%s/%s/%s",
}

// Remote returns a remote string that can be passed to git
func (gs *Git) Remote() string {
	return fmt.Sprintf(gitProtoFmts[gs.Scheme],
		gs.Host, gs.User, gs.Repo,
	)
}

// regular expressions for matching package uris
const (
	gitSSHExp     = `ssh://git@(?P<host>.+)/(?P<user>.+)/(?P<repo>.+).git`
	githubSlugExp = `github.com/(?P<user>[-_a-zA-Z0-9]+)/(?P<repo>[-_a-zA-Z0-9]+)`
)

var (
	gitSSHRegex                   = regexp.MustCompile(gitSSHExp)
	gitSSHWithVersionRegex        = regexp.MustCompile(gitSSHExp + `@(?P<version>.*)`)
	gitSSHWithPathRegex           = regexp.MustCompile(gitSSHExp + `/(?P<subdir>.*)`)
	gitSSHWithPathAndVersionRegex = regexp.MustCompile(gitSSHExp + `/(?P<subdir>.*)@(?P<version>.*)`)

	githubSlugRegex                   = regexp.MustCompile(githubSlugExp)
	githubSlugWithVersionRegex        = regexp.MustCompile(githubSlugExp + `@(?P<version>.*)`)
	githubSlugWithPathRegex           = regexp.MustCompile(githubSlugExp + `/(?P<subdir>.*)`)
	githubSlugWithPathAndVersionRegex = regexp.MustCompile(githubSlugExp + `/(?P<subdir>.*)@(?P<version>.*)`)
)

func parseGit(uri string) *Dependency {
	var d = Dependency{
		Version: "master",
		Source:  Source{},
	}
	var gs *Git
	var version string

	switch {
	case githubSlugRegex.MatchString(uri):
		gs, version = parseGitHub(uri)
	case gitSSHRegex.MatchString(uri):
		gs, version = parseGitSSH(uri)
	default:
		return nil
	}

	if gs.Subdir != "" {
		gs.Subdir = "/" + gs.Subdir
	}

	d.Source.GitSource = gs
	if version != "" {
		d.Version = version
	}
	return &d
}

func parseGitSSH(p string) (gs *Git, version string) {
	gs, version = match(p, []*regexp.Regexp{
		gitSSHWithPathAndVersionRegex,
		gitSSHWithPathRegex,
		gitSSHWithVersionRegex,
		gitSSHRegex,
	})

	gs.Scheme = GitSchemeSSH
	return gs, version
}

func parseGitHub(p string) (gs *Git, version string) {
	gs, version = match(p, []*regexp.Regexp{
		githubSlugWithPathAndVersionRegex,
		githubSlugWithPathRegex,
		githubSlugWithVersionRegex,
		githubSlugRegex,
	})

	gs.Scheme = GitSchemeHTTPS
	gs.Host = "github.com"
	return gs, version
}

func match(p string, exps []*regexp.Regexp) (gs *Git, version string) {
	gs = &Git{}
	for _, e := range exps {
		if !e.MatchString(p) {
			continue
		}

		matches := reSubMatchMap(e, p)
		gs.Host = matches["host"]
		gs.User = matches["user"]
		gs.Repo = matches["repo"]

		if sd, ok := matches["subdir"]; ok {
			gs.Subdir = sd
		}

		return gs, matches["version"]
	}
	return gs, ""
}

func reSubMatchMap(r *regexp.Regexp, str string) map[string]string {
	match := r.FindStringSubmatch(str)
	subMatchMap := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}
