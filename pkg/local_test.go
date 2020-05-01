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
