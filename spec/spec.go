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

package spec

import (
	"encoding/json"
	"sort"
)

type JsonnetFile struct {
	Dependencies map[string]Dependency
}

type jsonFile struct {
	Dependencies []Dependency `json:"dependencies"`
}

func (jf *JsonnetFile) UnmarshalJSON(data []byte) error {
	var s jsonFile
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	jf.Dependencies = make(map[string]Dependency)
	for _, d := range s.Dependencies {
		jf.Dependencies[d.Name] = d
	}
	return nil
}

func (jf JsonnetFile) MarshalJSON() ([]byte, error) {
	var s jsonFile
	for _, d := range jf.Dependencies {
		s.Dependencies = append(s.Dependencies, d)
	}


	return json.Marshal(s)
}

type Dependency struct {
	Name      string `json:"name"`
	Source    Source `json:"source"`
	Version   string `json:"version"`
	Sum       string `json:"sum,omitempty"`
	DepSource string `json:"-"`
}

type Source struct {
	GitSource   *GitSource   `json:"git,omitempty"`
	LocalSource *LocalSource `json:"local,omitempty"`
}

type GitSource struct {
	Remote string `json:"remote"`
	Subdir string `json:"subdir"`
}

type LocalSource struct {
	Directory string `json:"directory"`
}
