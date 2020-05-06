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
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

func TestLocalInstall(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	vendorDir, err := ioutil.TempDir(cwd, "vendor")
	assert.NoError(t, err)
	defer os.RemoveAll(vendorDir)

	pkgDir, err := ioutil.TempDir(cwd, "foo")
	assert.NoError(t, err)
	defer os.RemoveAll(pkgDir)

	relPath, err := filepath.Rel(cwd, pkgDir)
	assert.NoError(t, err)

	p := NewLocalPackage(&deps.Local{Directory: relPath})
	lockVersion, err := p.Install(context.TODO(), "foo", vendorDir, "v1.0")
	assert.NoError(t, err)
	assert.Empty(t, lockVersion)
}

func TestLocalInstallSourceNotFound(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	vendorDir, err := ioutil.TempDir(cwd, "vendor")
	assert.NoError(t, err)
	defer os.RemoveAll(vendorDir)

	relPath := "foo"
	p := NewLocalPackage(&deps.Local{Directory: relPath})
	lockVersion, err := p.Install(context.TODO(), "foo", vendorDir, "v1.0")
	assert.Error(t, err)
	assert.Empty(t, lockVersion)
}

func TestLocalInstallTargetDoesNotExist(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	pkgDir, err := ioutil.TempDir(cwd, "foo")
	assert.NoError(t, err)
	defer os.RemoveAll(pkgDir)

	relPath, err := filepath.Rel(cwd, pkgDir)
	assert.NoError(t, err)

	p := NewLocalPackage(&deps.Local{Directory: relPath})
	lockVersion, err := p.Install(context.TODO(), "foo", "vendor", "v1.0")
	assert.Error(t, err)
	assert.Empty(t, lockVersion)
}

func TestLocalInstallSourceAndTargetDoNotExist(t *testing.T) {
	p := NewLocalPackage(&deps.Local{Directory: "foo"})
	lockVersion, err := p.Install(context.TODO(), "foo", "bar", "v1.0")
	assert.Error(t, err)
	assert.Empty(t, lockVersion)
}
