// +build go1.12

package main

import (
	"os"
	"path/filepath"
	"runtime/debug"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func newApp() *kingpin.Application {
	a := kingpin.New(filepath.Base(os.Args[0]), "A jsonnet package manager")
	d, ok := debug.ReadBuildInfo()
	if ok {
		return a.Version(d.Main.Version)
	}
	return a
}
