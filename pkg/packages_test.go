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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/stretchr/testify/assert"
)

const NotExist = "/this/does/not/exist"

func TestInsertDependency(t *testing.T) {
	deps := []spec.Dependency{{Name: "test1", Version: "latest"}}
	dep := spec.Dependency{Name: "test2", Version: "latest"}

	res, err := insertDependency(deps, dep)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 2 {
		t.Fatal("Incorrectly inserted")
	}
}

func TestFileExists(t *testing.T) {
	{
		exists, err := FileExists(NotExist)
		assert.False(t, exists)
		assert.Nil(t, err)
	}
	{
		tempFile, err := ioutil.TempFile("", "jb-exists")
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			err := os.Remove(tempFile.Name())
			assert.Nil(t, err)
		}()

		exists, err := FileExists(tempFile.Name())
		assert.True(t, exists)
		assert.Nil(t, err)
	}
}

func TestLoadJsonnetfile(t *testing.T) {
	empty := spec.JsonnetFile{}

	jsonnetfileContent := `{
    "dependencies": [
        {
            "name": "foobar",
            "source": {
                "git": {
                    "remote": "https://github.com/foobar/foobar",
                    "subdir": ""
                }
            },
            "version": "master"
        }
    ]
}
`
	jsonnetFileExpected := spec.JsonnetFile{
		Dependencies: []spec.Dependency{{
			Name: "foobar",
			Source: spec.Source{
				GitSource: &spec.GitSource{
					Remote: "https://github.com/foobar/foobar",
					Subdir: "",
				},
			},
			Version:   "master",
			DepSource: "",
		}},
	}

	{
		jf, err := LoadJsonnetfile(NotExist)
		assert.Equal(t, empty, jf)
		assert.Error(t, err)
	}
	{
		tempDir, err := ioutil.TempDir("", "jb-load-jsonnetfile")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := os.RemoveAll(tempDir)
			assert.Nil(t, err)
		}()

		tempFile := filepath.Join(tempDir, jsonnetfile.File)
		err = ioutil.WriteFile(tempFile, []byte(`{}`), os.ModePerm)
		assert.Nil(t, err)

		jf, err := LoadJsonnetfile(tempFile)
		assert.Nil(t, err)
		assert.Equal(t, empty, jf)
	}
	{
		tempDir, err := ioutil.TempDir("", "jb-load-jsonnetfile")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			err := os.RemoveAll(tempDir)
			assert.Nil(t, err)
		}()

		tempFile := filepath.Join(tempDir, jsonnetfile.File)
		err = ioutil.WriteFile(tempFile, []byte(jsonnetfileContent), os.ModePerm)
		assert.Nil(t, err)

		jf, err := LoadJsonnetfile(tempFile)
		assert.Nil(t, err)
		assert.Equal(t, jsonnetFileExpected, jf)
	}
}
