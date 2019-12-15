package deps

import (
	"os"
	"path/filepath"
)

type Dependency struct {
	Source  Source `json:"source"`
	Version string `json:"version"`
	Sum     string `json:"sum,omitempty"`
}

type Source struct {
	GitSource   *Git   `json:"git,omitempty"`
	LocalSource *Local `json:"local,omitempty"`
}

type Local struct {
	Directory string `json:"directory"`
}

func Parse(dir, uri string) *Dependency {
	if uri == "" {
		return nil
	}

	if d := parseGit(uri); d != nil {
		return d
	}

	return parseLocal(dir, uri)
}

func (d *Dependency) Name() string {
	return d.Source.Name()
}

func (s Source) Name() string {
	switch {
	case s.GitSource != nil:
		return s.GitSource.Name()
	case s.LocalSource != nil:
		return filepath.Base(s.LocalSource.Directory)
	default:
		return ""
	}
}

func parseLocal(dir, p string) *Dependency {
	clean := filepath.Clean(p)
	abs := filepath.Join(dir, clean)

	info, err := os.Stat(abs)
	if err != nil {
		return nil
	}

	if !info.IsDir() {
		return nil
	}

	return &Dependency{
		Source: Source{
			LocalSource: &Local{
				Directory: clean,
			},
		},
		Version: "",
	}
}
