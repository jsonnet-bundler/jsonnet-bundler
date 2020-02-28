package spec

import (
	"path/filepath"

	v0 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v0"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

func FromV0(mv0 v0.JsonnetFile) (JsonnetFile, error) {
	m := New()
	m.LegacyImports = true

	for name, old := range mv0.Dependencies {
		var d deps.Dependency

		switch {
		case old.Source.GitSource != nil:
			d = *deps.Parse("", old.Source.GitSource.Remote)

			subdir := filepath.Clean("/" + old.Source.GitSource.Subdir)
			d.Source.GitSource.Subdir = subdir

		case old.Source.LocalSource != nil:
			d = *deps.Parse("", old.Source.LocalSource.Directory)
		}

		d.Sum = old.Sum
		d.Version = old.Version
		d.LegacyNameCompat = name

		m.Dependencies[d.Name()] = d
	}

	return m, nil
}
