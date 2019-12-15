package deps

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	GitSchemeSSH   = "ssh://git@"
	GitSchemeHTTPS = "https://"
)

type Git struct {
	Scheme string

	Host   string
	User   string
	Repo   string
	Subdir string
}

type jsonGit struct {
	Remote string `json:"remote"`
	Subdir string `json:"subdir"`
}

func (gs *Git) MarshalJSON() ([]byte, error) {
	j := jsonGit{
		Remote: gs.Remote(),
		Subdir: strings.TrimPrefix(gs.Subdir, "/"),
	}
	return json.Marshal(j)
}

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

var gitProtoFmts = map[string]string{
	GitSchemeSSH:   GitSchemeSSH + "%s:%s/%s.git",
	GitSchemeHTTPS: GitSchemeHTTPS + "%s/%s/%s",
}

// Remote returns a remote string that can be passed to git
func (gs *Git) Remote() string {
	return fmt.Sprintf(gitProtoFmts[gs.Scheme],
		gs.Host, gs.User, gs.Repo,
	)
}

var (
	gitSSHRegex                   = regexp.MustCompile(`git\+ssh://git@([^:]+):([^/]+)/([^/]+).git`)
	gitSSHWithVersionRegex        = regexp.MustCompile(`git\+ssh://git@([^:]+):([^/]+)/([^/]+).git@(.*)`)
	gitSSHWithPathRegex           = regexp.MustCompile(`git\+ssh://git@([^:]+):([^/]+)/([^/]+).git/(.*)`)
	gitSSHWithPathAndVersionRegex = regexp.MustCompile(`git\+ssh://git@([^:]+):([^/]+)/([^/]+).git/(.*)@(.*)`)

	githubSlugRegex                   = regexp.MustCompile("github.com/([-_a-zA-Z0-9]+)/([-_a-zA-Z0-9]+)")
	githubSlugWithVersionRegex        = regexp.MustCompile("github.com/([-_a-zA-Z0-9]+)/([-_a-zA-Z0-9]+)@(.*)")
	githubSlugWithPathRegex           = regexp.MustCompile("github.com/([-_a-zA-Z0-9]+)/([-_a-zA-Z0-9]+)/(.*)")
	githubSlugWithPathAndVersionRegex = regexp.MustCompile("github.com/([-_a-zA-Z0-9]+)/([-_a-zA-Z0-9]+)/(.*)@(.*)")
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
	gs = &Git{
		Scheme: GitSchemeSSH,
	}

	switch {
	case gitSSHWithPathAndVersionRegex.MatchString(p):
		matches := gitSSHWithPathAndVersionRegex.FindStringSubmatch(p)
		gs.Host = matches[1]
		gs.User = matches[2]
		gs.Repo = matches[3]
		gs.Subdir = matches[4]
		version = matches[5]
	case gitSSHWithPathRegex.MatchString(p):
		matches := gitSSHWithPathRegex.FindStringSubmatch(p)
		gs.Host = matches[1]
		gs.User = matches[2]
		gs.Repo = matches[3]
		gs.Subdir = matches[4]
	case gitSSHWithVersionRegex.MatchString(p):
		matches := gitSSHWithVersionRegex.FindStringSubmatch(p)
		gs.Host = matches[1]
		gs.User = matches[2]
		gs.Repo = matches[3]
		version = matches[4]
	default:
		matches := gitSSHRegex.FindStringSubmatch(p)
		gs.Host = matches[1]
		gs.User = matches[2]
		gs.Repo = matches[3]
	}

	return gs, version
}
func parseGitHub(p string) (gs *Git, version string) {
	gs = &Git{
		Scheme: GitSchemeHTTPS,
		Host:   "github.com",
	}

	if githubSlugWithPathRegex.MatchString(p) {
		if githubSlugWithPathAndVersionRegex.MatchString(p) {
			matches := githubSlugWithPathAndVersionRegex.FindStringSubmatch(p)
			gs.User = matches[1]
			gs.Repo = matches[2]
			gs.Subdir = matches[3]
			version = matches[4]
		} else {
			matches := githubSlugWithPathRegex.FindStringSubmatch(p)
			gs.User = matches[1]
			gs.Repo = matches[2]
			gs.Subdir = matches[3]
		}
	} else {
		if githubSlugWithVersionRegex.MatchString(p) {
			matches := githubSlugWithVersionRegex.FindStringSubmatch(p)
			gs.User = matches[1]
			gs.Repo = matches[2]
			version = matches[3]
		} else {
			matches := githubSlugRegex.FindStringSubmatch(p)
			gs.User = matches[1]
			gs.Repo = matches[2]
		}
	}

	return gs, version
}
