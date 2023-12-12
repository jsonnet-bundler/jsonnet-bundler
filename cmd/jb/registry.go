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

package main

import (
	"context"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg"
	"gopkg.in/alecthomas/kingpin.v2"
)

func registryAddCommand(name, description, url, filename string) int {
	registry := pkg.NewGitRegistry(name, description, url, filename)

	for _, r := range pkg.Registries.Entries {
		if r.Name == name {
			kingpin.Fatalf("Registry %s already exists", name)
		}
	}

	pkg.Registries.Entries = append(pkg.Registries.Entries, *registry)

	err := registry.Init(context.TODO())
	if err != nil {
		kingpin.FatalIfError(err, "could not init registry")
	}
	err = registry.Update(context.TODO())
	if err != nil {
		kingpin.FatalIfError(err, "could not update registry")
	}
	return 0
}

func registryRemoveCommand(name string) int {
	var notRemoved []pkg.GitRegistry
	var err error

	for _, r := range pkg.Registries.Entries {
		if r.Name != name {
			notRemoved = append(notRemoved, r)
		} else {
			err = r.CleanCache()
			if err != nil {
				kingpin.FatalIfError(err, "could not clean cache files of registry %s", r.Name)
			}
		}
	}
	pkg.Registries.Entries = notRemoved
	pkg.Registries.SaveRegistries()

	return 0
}

func registryUpdateCommand() int {
	err := pkg.UpdateRegistries(context.TODO())
	if err != nil {
		kingpin.FatalIfError(err, "could not update registry")
	}
	return 0
}

func registryListCommand() int {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
	pkg.PrintHeader(w, []string{"Name", "Description", "Source"})

	for _, registry := range pkg.Registries.Entries {
		pkg.PrintRow(w, []string{registry.Name, registry.Description, registry.Source})
	}
	w.Flush()
	return 0
}

func registrySearchCommand(query string, details, allVersions bool) int {
	results, err := pkg.SearchPackage(context.TODO(), query)
	if err != nil {
		kingpin.FatalIfError(err, "could not search package")
	}

	var w *tabwriter.Writer

	if details {
		w = tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
		pkg.PrintHeader(w, []string{"Name", "Description", "Registry", "Versions"})
	} else if allVersions {
		w = tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
		pkg.PrintHeader(w, []string{"Name", "Version", "Url"})
	} else {
		w = tabwriter.NewWriter(os.Stdout, 1, 1, 5, ' ', 0)
		pkg.PrintHeader(w, []string{"Name", "Url"})
	}

	for registryName, packages := range results {
		if len(packages) > 0 {
			for _, p := range packages {
				sort.Sort(p.Versions)
				if details {
					versions := []string{}
					for _, v := range p.Versions {
						versions = append(versions, v.Version)
					}
					pkg.PrintRow(w, []string{p.Name, p.Description, registryName, strings.Join(versions, ",")})
				} else if allVersions {
					for _, v := range p.Versions {
						pkg.PrintRow(w, []string{p.Name, v.Version, v.Source})
					}
				} else {
					pkg.PrintRow(w, []string{p.Name, p.Versions[len(p.Versions)-1].Source})
				}
			}
		}
	}
	w.Flush()
	return 0
}
