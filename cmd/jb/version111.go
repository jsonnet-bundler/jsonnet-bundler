// +build !go1.12

package main

import (
	"os"
	"path/filepath"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func newApp() *kingpin.Application {
	a := kingpin.New(filepath.Base(os.Args[0]), "A jsonnet package manager")
	return a
}
