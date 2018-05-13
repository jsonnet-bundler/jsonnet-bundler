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

	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
)

func TestInsert(t *testing.T) {
	deps := []*spec.Dependency{&spec.Dependency{Name: "test1", Version: "latest"}}
	dep := &spec.Dependency{Name: "test2", Version: "latest"}

	res, err := insertDependency(deps, dep)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 2 {
		t.Fatal("Incorrectly inserted")
	}
}
