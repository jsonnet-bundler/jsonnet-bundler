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
package rewrite

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

const sample = `
(import "k.libsonnet") + // not vendored
(import "ksonnet/abc.jsonnet") + // prefix of next
(import "ksonnet.beta.4/k.libsonnet") + // normal import
(import "github.com/ksonnet/ksonnet/def.jsonnet") + // already absolute
(import "prometheus/mixin/whatever/abc.libsonnet") + // nested
(import "mylib/foo.libsonnet") + // not managed by jb
// completely unrelated line:
[ "nice" ]
`

const want = `
(import "k.libsonnet") + // not vendored
(import "github.com/ksonnet/ksonnet/abc.jsonnet") + // prefix of next
(import "github.com/ksonnet/ksonnet-lib/ksonnet.beta.4/k.libsonnet") + // normal import
(import "github.com/ksonnet/ksonnet/def.jsonnet") + // already absolute
(import "github.com/prometheus/prometheus/mixin/whatever/abc.libsonnet") + // nested
(import "mylib/foo.libsonnet") + // not managed by jb
// completely unrelated line:
[ "nice" ]
`

func TestRewrite(t *testing.T) {
	testRewriteWithJsonnetHome(t, "vendor")
}

func TestRewriteCustomJsonnetHome(t *testing.T) {
	testRewriteWithJsonnetHome(t, "custom-vendor-dir")
}

func TestRewriteDeepCustomJsonnetHome(t *testing.T) {
	testRewriteWithJsonnetHome(t, "custom/vendor/dir")
}

func testRewriteWithJsonnetHome(t *testing.T, jsonnetHome string) {
	dir, err := ioutil.TempDir("", "jbrewrite")
	require.Nil(t, err)
	defer os.RemoveAll(dir)

	name := filepath.Join(dir, "test.jsonnet")
	err = ioutil.WriteFile(name, []byte(sample), 0644)
	require.Nil(t, err)

	jsonnetHome = filepath.Join(dir, jsonnetHome)
	err = os.MkdirAll(jsonnetHome, os.ModePerm)
	require.Nil(t, err)

	err = Rewrite(dir, jsonnetHome, locks)
	require.Nil(t, err)

	content, err := ioutil.ReadFile(name)
	require.Nil(t, err)

	assert.Equal(t, want, string(content))
}

var locks = map[string]deps.Dependency{
	"ksonnet":        *deps.Parse("", "github.com/ksonnet/ksonnet"),
	"ksonnet.beta.4": *deps.Parse("", "github.com/ksonnet/ksonnet-lib/ksonnet.beta.4"),
	"prometheus":     *deps.Parse("", "github.com/prometheus/prometheus"),
}
