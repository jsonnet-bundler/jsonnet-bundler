package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg"
	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"gopkg.in/alecthomas/kingpin.v2"
)

func installCommand(dir, jsonnetHome string, urls ...*url.URL) int {
	if dir == "" {
		dir = "."
	}

	filename, isLock, err := jsonnetfile.Choose(dir)
	if err != nil {
		kingpin.Fatalf("failed to choose jsonnetfile: %v", err)
		return 1
	}

	jsonnetFile, err := jsonnetfile.Load(filename)
	if err != nil {
		kingpin.Fatalf("failed to load jsonnetfile: %v", err)
		return 1
	}

	if len(urls) > 0 {
		for _, url := range urls {
			// install package specified in command
			// $ jsonnetpkg install ksonnet git@github.com:ksonnet/ksonnet-lib
			// $ jsonnetpkg install grafonnet git@github.com:grafana/grafonnet-lib grafonnet
			// $ jsonnetpkg install github.com/grafana/grafonnet-lib/grafonnet
			//
			// github.com/(slug)/(dir)

			urlString := url.String()
			newDep := parseDepedency(urlString)
			if newDep == nil {
				kingpin.Errorf("ignoring unrecognized url: %s", url)
				continue
			}

			oldDeps := jsonnetFile.Dependencies
			newDeps := []spec.Dependency{}
			oldDepReplaced := false
			for _, d := range oldDeps {
				if d.Name == newDep.Name {
					newDeps = append(newDeps, *newDep)
					oldDepReplaced = true
				} else {
					newDeps = append(newDeps, d)
				}
			}

			if !oldDepReplaced {
				newDeps = append(newDeps, *newDep)
			}

			jsonnetFile.Dependencies = newDeps
		}
	}

	srcPath := filepath.Join(jsonnetHome)
	err = os.MkdirAll(srcPath, os.ModePerm)
	if err != nil {
		kingpin.Fatalf("failed to create jsonnet home path: %v", err)
		return 3
	}

	lock, err := pkg.Install(context.TODO(), isLock, filename, jsonnetFile, jsonnetHome)
	if err != nil {
		kingpin.Fatalf("failed to install: %v", err)
		return 3
	}

	// If installing from lock file there is no need to write any files back.
	if !isLock {
		b, err := json.MarshalIndent(jsonnetFile, "", "    ")
		if err != nil {
			kingpin.Fatalf("failed to encode jsonnet file: %v", err)
			return 3
		}
		b = append(b, []byte("\n")...)

		err = ioutil.WriteFile(filepath.Join(dir, jsonnetfile.File), b, 0644)
		if err != nil {
			kingpin.Fatalf("failed to write jsonnet file: %v", err)
			return 3
		}

		b, err = json.MarshalIndent(lock, "", "    ")
		if err != nil {
			kingpin.Fatalf("failed to encode jsonnet file: %v", err)
			return 3
		}
		b = append(b, []byte("\n")...)

		err = ioutil.WriteFile(filepath.Join(dir, jsonnetfile.LockFile), b, 0644)
		if err != nil {
			kingpin.Fatalf("failed to write lock file: %v", err)
			return 3
		}
	}

	return 0
}
