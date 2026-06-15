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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

// OCIPlainHTTP forces pulls over plain HTTP instead of HTTPS. It is intended
// for local or insecure registries, mirroring the `oras --plain-http` flag.
var OCIPlainHTTP = false

type OCIPackage struct {
	Source *deps.OCI
}

func NewOCIPackage(source *deps.OCI) Interface {
	return &OCIPackage{
		Source: source,
	}
}

// Install pulls the artifact at the requested version (a tag or digest) from
// the OCI registry, reconstructs its files in the vendor directory and returns
// the resolved manifest digest as the lock version. Credentials are resolved
// the same way the `oras` CLI does, from the docker config, falling back to
// anonymous pulls.
func (p *OCIPackage) Install(ctx context.Context, name, dir, version string) (string, error) {
	destPath := path.Join(dir, name)

	tmpDir, err := os.MkdirTemp(filepath.Join(dir, ".tmp"), "oci")
	if err != nil {
		return "", errors.Wrap(err, "failed to create tmp dir")
	}
	defer os.RemoveAll(tmpDir)

	contentDir := filepath.Join(tmpDir, "content")
	fs, err := file.New(contentDir)
	if err != nil {
		return "", errors.Wrap(err, "failed to create file store")
	}
	defer fs.Close()

	repo, err := remote.NewRepository(p.Source.Repo)
	if err != nil {
		return "", errors.Wrap(err, "invalid oci repository")
	}
	repo.PlainHTTP = OCIPlainHTTP

	credStore, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err != nil {
		return "", errors.Wrap(err, "failed to load docker credentials")
	}
	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credentials.Credential(credStore),
	}

	// A digest reference is joined with "@", a tag with ":".
	sep := ":"
	if strings.Contains(version, ":") {
		sep = "@"
	}
	color.Cyan("PULL oci://%s%s%s", p.Source.Repo, sep, version)

	desc, err := oras.Copy(ctx, repo, version, fs, version, oras.DefaultCopyOptions)
	if err != nil {
		return "", errors.Wrap(err, "failed to pull oci artifact")
	}

	// Release the store's handles before moving the extracted tree.
	if err := fs.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close file store")
	}

	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
		return "", errors.Wrap(err, "failed to create parent path")
	}

	if err := os.RemoveAll(destPath); err != nil {
		return "", errors.Wrap(err, "failed to clean previous destination path")
	}

	if err := os.Rename(contentDir, destPath); err != nil {
		return "", errors.Wrap(err, "failed to move package")
	}

	return desc.Digest.String(), nil
}
