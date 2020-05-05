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

// +build integration

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RepoState describes a point in time of a repository
type RepoState struct {
	File string
	Lock string
}

// FilePath is the path to the jsonnetfile.json
func (rs RepoState) FilePath(dir string) string {
	return filepath.Join(dir, jsonnetfile.File)
}

// LockPath is the path to the jsonnetfile.lock.json
func (rs RepoState) LockPath(dir string) string {
	return filepath.Join(dir, jsonnetfile.LockFile)
}

// Write writes this state to dir
func (rs RepoState) Write(dir string) error {
	if err := ioutil.WriteFile(rs.FilePath(dir), []byte(rs.File), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(rs.LockPath(dir), []byte(rs.Lock), 0644); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "vendor/"), os.ModePerm); err != nil {
		return err
	}
	return nil
}

// Assert checks that dir matches this state
func (rs RepoState) Assert(t *testing.T, dir string) {
	file, err := ioutil.ReadFile(rs.FilePath(dir))
	require.NoError(t, err)
	assert.JSONEq(t, rs.File, string(file))

	lock, err := ioutil.ReadFile(rs.LockPath(dir))
	require.NoError(t, err)
	assert.JSONEq(t, rs.Lock, string(lock))
}

// UpdateCase is a testcase for jb update
type UpdateCase struct {
	name   string
	uris   []string
	before *RepoState
	after  *RepoState
}

func (u UpdateCase) Run(t *testing.T) {
	dir, err := ioutil.TempDir("", u.name)
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	if u.before == nil {
		initCommand(dir)
	} else {
		err = u.before.Write(dir)
		require.NoError(t, err)
	}

	ret := updateCommand(dir, "vendor", u.uris)
	assert.Equal(t, ret, 0)

	if u.after != nil {
		u.after.Assert(t, dir)
	}
}

func TestUpdate(t *testing.T) {
	cases := []UpdateCase{
		{
			name: "simple",
			uris: []string{}, // no uris
			before: &RepoState{
				File: `{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/jsonnet-bundler/frozen-lib","subdir":""}},"version":"master"}],"legacyImports":true}`,
				Lock: `{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/jsonnet-bundler/frozen-lib","subdir":""}},"version":"9f40207f668e382b706e1822f2d46ce2cd0a57cc","sum":"qUJDskVRtmkTms2udvFpLi1t5YKVbGmMSyiZnPjXsMo="}],"legacyImports":false}`,
			},
			after: &RepoState{
				File: `{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/jsonnet-bundler/frozen-lib","subdir":""}},"version":"master"}],"legacyImports":true}`,
				Lock: `{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/jsonnet-bundler/frozen-lib","subdir":""}},"version":"ed7c1aff9e10d3b42fb130446d495f1c769ecd7b","sum":"OraOcUvDIx9Eikaihi8XsRNRsVehO75Ek35im/jYoSA="}],"legacyImports":false}`,
			},
		},
		{
			name: "single",
			uris: []string{"github.com/jsonnet-bundler/frozen-lib"},
			before: &RepoState{
				File: `{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/grafana/jsonnet-libs","subdir":"ksonnet-util"}},"version":"master"},{"source":{"git":{"remote":"https://github.com/jsonnet-bundler/frozen-lib","subdir":""}},"version":"master"}],"legacyImports":true}`,
				Lock: `{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/grafana/jsonnet-libs","subdir":"ksonnet-util"}},"version":"610b00d219d0a6f3d833dd44e4bb0deda2429da0","sum":"XdIrw3m7I8fJ3CL9eR8LtuYcanf2QK78n4H4OBBOADc="},{"source":{"git":{"remote":"https://github.com/jsonnet-bundler/frozen-lib","subdir":""}},"version":"9f40207f668e382b706e1822f2d46ce2cd0a57cc","sum":"qUJDskVRtmkTms2udvFpLi1t5YKVbGmMSyiZnPjXsMo="}],"legacyImports":false}`,
			},
			after: &RepoState{
				File: `{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/grafana/jsonnet-libs","subdir":"ksonnet-util"}},"version":"master"},{"source":{"git":{"remote":"https://github.com/jsonnet-bundler/frozen-lib","subdir":""}},"version":"master"}],"legacyImports":true}`,
				Lock: `{"version":1,"dependencies":[{"source":{"git":{"remote":"https://github.com/grafana/jsonnet-libs","subdir":"ksonnet-util"}},"version":"610b00d219d0a6f3d833dd44e4bb0deda2429da0","sum":"XdIrw3m7I8fJ3CL9eR8LtuYcanf2QK78n4H4OBBOADc="},{"source":{"git":{"remote":"https://github.com/jsonnet-bundler/frozen-lib","subdir":""}},"version":"ed7c1aff9e10d3b42fb130446d495f1c769ecd7b","sum":"OraOcUvDIx9Eikaihi8XsRNRsVehO75Ek35im/jYoSA="}],"legacyImports":false}`,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, c.Run)
	}
}
