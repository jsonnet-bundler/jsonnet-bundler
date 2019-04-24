package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg"
	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"gopkg.in/alecthomas/kingpin.v2"
)

func initCommand(dir string) int {
	exists, err := pkg.FileExists(jsonnetfile.File)
	if err != nil {
		kingpin.Errorf("Failed to check for jsonnetfile.json: %v", err)
		return 1
	}

	if exists {
		kingpin.Errorf("jsonnetfile.json already exists")
		return 1
	}

	filename := filepath.Join(dir, jsonnetfile.File)

	if err := ioutil.WriteFile(filename, []byte("{}\n"), 0644); err != nil {
		kingpin.Errorf("Failed to write new jsonnetfile.json: %v", err)
		return 1
	}

	return 0
}
