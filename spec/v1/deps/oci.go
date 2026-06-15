// Copyright 2018 jsonnet-bundler authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package deps

import (
	"path/filepath"
	"strings"

	"oras.land/oras-go/v2/registry"
)

const OCIScheme = "oci://"

// OCI holds the information required to pull a package from an OCI registry
type OCI struct {
	// Repo is the fully qualified artifact reference without a tag or digest,
	// e.g. "registry.example.com/namespace/name".
	Repo string `json:"repository"`
}

// Name returns the artifact reference, used as the on-disk vendor path
// (example.com/namespace/name).
func (o *OCI) Name() string {
	return o.Repo
}

// LegacyName returns the last element of the artifact path.
func (o *OCI) LegacyName() string {
	return filepath.Base(o.Repo)
}

func parseOCI(uri string) *Dependency {
	if !strings.HasPrefix(uri, OCIScheme) {
		return nil
	}

	ref, err := registry.ParseReference(strings.TrimPrefix(uri, OCIScheme))
	if err != nil {
		return nil
	}

	d := &Dependency{
		Version: "latest",
		Source: Source{
			OCISource: &OCI{
				Repo: ref.Registry + "/" + ref.Repository,
			},
		},
	}

	// ref.Reference holds the tag or digest when present.
	if ref.Reference != "" {
		d.Version = ref.Reference
	}

	return d
}
