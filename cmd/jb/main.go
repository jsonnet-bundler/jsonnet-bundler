/*
Copyright 2018 jsonnet-bundler authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/pkg/errors"
)

const (
	installSubcommand = "install"
	initSubcommand    = "init"
	basePath          = ".jsonnetpkg"
	srcDirName        = "src"
)

var (
	availableSubcommands = []string{
		initSubcommand,
		installSubcommand,
	}
	githubSlugRegex                   = regexp.MustCompile("github.com/(.*)/(.*)")
	githubSlugWithVersionRegex        = regexp.MustCompile("github.com/(.*)/(.*)@(.*)")
	githubSlugWithPathRegex           = regexp.MustCompile("github.com/(.*)/(.*)/(.*)")
	githubSlugWithPathAndVersionRegex = regexp.MustCompile("github.com/(.*)/(.*)/(.*)@(.*)")
)

type config struct {
	JsonnetHome string
}

func Main() int {
	cfg := config{}

	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.StringVar(&cfg.JsonnetHome, "jsonnetpkg-home", "vendor", "The directory used to cache packages in.")
	flagset.Parse(os.Args[1:])

	subcommand := "install"
	args := flagset.Args()
	if len(args) >= 1 {
		subcommand = args[0]
	}

	err := RunSubcommand(context.TODO(), cfg, subcommand, args[1:])
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		return 1
	}

	return 0
}

func RunSubcommand(ctx context.Context, cfg config, subcommand string, args []string) error {
	switch subcommand {
	case initSubcommand:
		return ioutil.WriteFile(pkg.JsonnetFile, []byte("{}"), 0644)
	case installSubcommand:
		m, err := pkg.LoadJsonnetfile(pkg.JsonnetFile)
		if err != nil {
			return errors.Wrap(err, "failed to load jsonnetfile")
		}

		if len(args) == 1 {
			// install package specified in command
			// $ jsonnetpkg install ksonnet git@github.com:ksonnet/ksonnet-lib
			// $ jsonnetpkg install grafonnet git@github.com:grafana/grafonnet-lib grafonnet
			// $ jsonnetpkg install github.com/grafana/grafonnet-lib/grafonnet
			//
			// github.com/(slug)/(dir)

			if githubSlugRegex.MatchString(args[0]) {
				name := ""
				user := ""
				repo := ""
				subdir := ""
				version := "master"
				if githubSlugWithPathRegex.MatchString(args[0]) {
					if githubSlugWithPathAndVersionRegex.MatchString(args[0]) {
						matches := githubSlugWithPathAndVersionRegex.FindStringSubmatch(args[0])
						user = matches[1]
						repo = matches[2]
						subdir = matches[3]
						version = matches[4]
						name = path.Base(subdir)
					} else {
						matches := githubSlugWithPathRegex.FindStringSubmatch(args[0])
						user = matches[1]
						repo = matches[2]
						subdir = matches[3]
						name = path.Base(subdir)
					}
				} else {
					if githubSlugWithVersionRegex.MatchString(args[0]) {
						matches := githubSlugWithVersionRegex.FindStringSubmatch(args[0])
						user = matches[1]
						repo = matches[2]
						name = repo
						version = matches[3]
					} else {
						matches := githubSlugRegex.FindStringSubmatch(args[0])
						user = matches[1]
						repo = matches[2]
						name = repo
					}
				}

				newDep := spec.Dependency{
					Name: name,
					Source: spec.Source{
						GitSource: &spec.GitSource{
							Remote: fmt.Sprintf("git@github.com:%s/%s", user, repo),
							Subdir: subdir,
						},
					},
					Version: version,
				}
				oldDeps := m.Dependencies
				newDeps := []spec.Dependency{}
				oldDepReplaced := false
				for _, d := range oldDeps {
					if d.Name == newDep.Name {
						newDeps = append(newDeps, newDep)
						oldDepReplaced = true
					} else {
						newDeps = append(newDeps, d)
					}
				}

				if !oldDepReplaced {
					newDeps = append(newDeps, newDep)
				}

				m.Dependencies = newDeps
			}
		}

		srcPath := filepath.Join(cfg.JsonnetHome)
		err = os.MkdirAll(srcPath, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "failed to create jsonnet home path")
		}

		lock, err := pkg.Install(ctx, m, cfg.JsonnetHome)
		if err != nil {
			return errors.Wrap(err, "failed to install")
		}

		b, err := json.MarshalIndent(m, "", "    ")
		if err != nil {
			return errors.Wrap(err, "failed to encode jsonnet file")
		}

		err = ioutil.WriteFile(pkg.JsonnetFile, b, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to write jsonnet file")
		}

		b, err = json.MarshalIndent(lock, "", "    ")
		if err != nil {
			return errors.Wrap(err, "failed to encode jsonnet file")
		}

		err = ioutil.WriteFile(pkg.JsonnetLockFile, b, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to write lock file")
		}
	default:
		return fmt.Errorf("Subcommand \"%s\" not availble. Available subcommands: %s", subcommand, strings.Join(availableSubcommands, ", "))
	}

	return nil
}

func main() {
	os.Exit(Main())
}
