package jsonnetfile_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/stretchr/testify/assert"
)

const notExist = "/this/does/not/exist"

func TestChoose(t *testing.T) {
	testcases := []struct {
		Name             string
		Jsonnetfile      []byte
		JsonnetfileLock  []byte
		ExpectedFilename string
		ExpectedLock     bool
		ExpectedError    error
	}{{
		Name:             "NoFiles",
		ExpectedFilename: "",
		ExpectedLock:     false,
		ExpectedError:    jsonnetfile.ErrNoFile,
	}, {
		Name:             "Jsonnetfile",
		Jsonnetfile:      []byte(`{}`),
		ExpectedFilename: jsonnetfile.File,
		ExpectedLock:     false,
		ExpectedError:    nil,
	}, {
		Name:             "JsonnetfileLock",
		Jsonnetfile:      []byte(`{}`),
		JsonnetfileLock:  []byte(`{}`),
		ExpectedFilename: jsonnetfile.LockFile,
		ExpectedLock:     true,
		ExpectedError:    nil,
	}}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "jsonnetfile-choose")
			assert.Nil(t, err)
			defer os.Remove(dir)

			if tc.Jsonnetfile != nil {
				err := ioutil.WriteFile(filepath.Join(dir, jsonnetfile.File), tc.Jsonnetfile, os.ModePerm)
				assert.NoError(t, err)
			}
			if tc.JsonnetfileLock != nil {
				err := ioutil.WriteFile(filepath.Join(dir, jsonnetfile.LockFile), tc.JsonnetfileLock, os.ModePerm)
				assert.NoError(t, err)
			}

			filename, isLock, err := jsonnetfile.Choose(dir)

			assert.Equal(t, tc.ExpectedFilename, strings.TrimPrefix(filename, dir+"/"))
			assert.Equal(t, tc.ExpectedLock, isLock)

			if tc.ExpectedError != nil {
				assert.EqualError(t, err, tc.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
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
		jf, err := jsonnetfile.Load(notExist)
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

		jf, err := jsonnetfile.Load(tempFile)
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

		jf, err := jsonnetfile.Load(tempFile)
		assert.Nil(t, err)
		assert.Equal(t, jsonnetFileExpected, jf)
	}
}
