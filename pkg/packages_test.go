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

package pkg

import (
	"testing"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

func TestKnown(t *testing.T) {
	testDeps := deps.NewOrdered()
	testDeps.Set("ksonnet-lib", deps.Dependency{
		Source: deps.Source{
			GitSource: &deps.Git{
				Scheme: deps.GitSchemeHTTPS,
				Host:   "github.com",
				User:   "ksonnet",
				Repo:   "ksonnet-lib",
				Subdir: "/ksonnet.beta.4",
			},
		},
	})

	paths := []string{
		"github.com",
		"github.com/ksonnet",
		"github.com/ksonnet/ksonnet-lib",
		"github.com/ksonnet/ksonnet-lib/ksonnet.beta.4",
		"github.com/ksonnet/ksonnet-lib/ksonnet.beta.4/k.libsonnet",
		"github.com/ksonnet-util", // don't know that one
		"ksonnet.beta.4",          // the symlink
	}

	want := []string{
		"github.com",
		"github.com/ksonnet",
		"github.com/ksonnet/ksonnet",
		"github.com/ksonnet/ksonnet-lib",
		"github.com/ksonnet/ksonnet-lib/ksonnet.beta.4",
		"github.com/ksonnet/ksonnet-lib/ksonnet.beta.4/k.libsonnet",
	}

	w := make(map[string]bool)
	for _, k := range want {
		w[k] = true
	}

	for _, p := range paths {
		if known(testDeps, p) != w[p] {
			t.Fatalf("expected %s to be %v", p, w[p])
		}
	}
}

func TestCleanLegacyName(t *testing.T) {
	testList := func(name string) *deps.Ordered {
		l := deps.NewOrdered()
		l.Set("ksonnet-lib", deps.Dependency{
			LegacyNameCompat: name,
			Source: deps.Source{
				GitSource: &deps.Git{
					Scheme: deps.GitSchemeHTTPS,
					Host:   "github.com",
					User:   "ksonnet",
					Repo:   "ksonnet-lib",
					Subdir: "/ksonnet.beta.4",
				}},
		})
		return l
	}
	cases := map[string]bool{
		"ksonnet":        false,
		"ksonnet.beta.4": true,
	}

	for name, want := range cases {
		list := testList(name)
		CleanLegacyName(list)
		if (list.GetOrDefault("ksonnet-lib", deps.Dependency{}).LegacyNameCompat == "") != want {
			t.Fatalf("expected `%s` to be removed: %v", name, want)
		}
	}
}
