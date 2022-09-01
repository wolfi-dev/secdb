// Copyright 2022 Chainguard, Inc.
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

package types

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TODO(kaniini): Harmonize these types with alpine's secdb implementation
// and place the implementation in alpine-go.
type Secfixes map[string][]string

type Package struct {
	Name     string   `json:"name"`
	Secfixes Secfixes `json:"secfixes"`
}

type PackageEntry struct {
	Pkg Package `json:"pkg"`
}

type Database struct {
	Apkurl    string         `json:"apkurl"`
	Archs     []string       `json:"archs"`
	Reponame  string         `json:"reponame"`
	Urlprefix string         `json:"urlprefix"`
	Packages  []PackageEntry `json:"packages"`
}

type MelangePackage struct {
	Package struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Epoch   int    `yaml:"epoch"`
	} `yaml:"package"`
	Secfixes Secfixes `yaml:"secfixes"`
}

// Identity returns the package identity triple as apk-tools expects
// it to be, e.g. `[name]-[version]-r[epoch].`
func (mp MelangePackage) Identity() string {
	return fmt.Sprintf("%s-%s-r%d", mp.Package.Name, mp.Package.Version, mp.Package.Epoch)
}

func (mp MelangePackage) Entry() PackageEntry {
	return PackageEntry{
		Pkg: Package{
			Name:     mp.Package.Name,
			Secfixes: mp.Secfixes,
		},
	}
}

// LoadMelangePackage loads a Melange source package YAML file
// and extracts the Secfixes data from it.
func LoadMelangePackage(fileName string) (*MelangePackage, error) {
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	pkg := &MelangePackage{}
	if err := yaml.Unmarshal(fileData, pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}
