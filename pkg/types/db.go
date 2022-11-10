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
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

const (
	substitutionPackageName          = "${{package.name}}"
	substitutionPackageVersion       = "${{package.version}}"
	substitutionPackageEpoch         = "${{package.epoch}}"
	substitutionTargetsDestdir       = "${{targets.destdir}}"
	substitutionSubPkgDir            = "${{targets.subpkgdir}}"
	substitutionHostTripletGnu       = "${{host.triplet.gnu}}"
	substitutionHostTripletRust      = "${{host.triplet.rust}}"
	substitutionCrossTripletGnuGlibc = "${{cross.triplet.gnu.glibc}}"
	substitutionCrossTripletGnuMusl  = "${{cross.triplet.gnu.musl}}"
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

func substitutionReplacements() map[string]string {
	return map[string]string{
		substitutionPackageName:          "MELANGE_TEMP_REPLACEMENT_PACAKAGE_NAME",
		substitutionPackageVersion:       "MELANGE_TEMP_REPLACEMENT_PACAKAGE_VERSION",
		substitutionPackageEpoch:         "MELANGE_TEMP_REPLACEMENT_PACAKAGE_EPOCH",
		substitutionTargetsDestdir:       "MELANGE_TEMP_REPLACEMENT_DESTDIR",
		substitutionSubPkgDir:            "MELANGE_TEMP_REPLACEMENT_SUBPKGDIR",
		substitutionHostTripletGnu:       "MELANGE_TEMP_REPLACEMENT_HOST_TRIPLET_GNU",
		substitutionHostTripletRust:      "MELANGE_TEMP_REPLACEMENT_HOST_TRIPLET_RUST",
		substitutionCrossTripletGnuGlibc: "MELANGE_TEMP_REPLACEMENT_CROSS_TRIPLET_GNU_GLIBC",
		substitutionCrossTripletGnuMusl:  "MELANGE_TEMP_REPLACEMENT_CROSS_TRIPLET_GNU_MUSL",
	}
}

func applyTemplate(contents []byte) ([]byte, error) {
	// First, replace all protected pipeline templated vars temporarily
	// So that we can apply the Go template
	// We have to do this bc go templates doesn't support ignoring certain fields: https://github.com/golang/go/issues/31147

	sr := substitutionReplacements()

	protected := string(contents)
	for k, v := range sr {
		protected = strings.ReplaceAll(protected, k, v)
	}

	tmpl, err := template.New("").Funcs(sprig.TxtFuncMap()).Parse(protected)
	if err != nil {
		return nil, err
	}
	tmpl = tmpl.Option("missingkey=error")
	buf := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(buf, nil); err != nil {
		return nil, err
	}

	// Add the pipeline templating back in
	templateApplied := buf.String()
	for k, v := range sr {
		templateApplied = strings.ReplaceAll(templateApplied, v, k)
	}

	return []byte(templateApplied), nil
}

// LoadMelangePackage loads a Melange source package YAML file
// and extracts the Secfixes data from it.
func LoadMelangePackage(fileName string) (*MelangePackage, error) {
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	tmplData, err := applyTemplate(fileData)
	if err != nil {
		return nil, err
	}

	pkg := &MelangePackage{}
	if err := yaml.Unmarshal(tmplData, pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}
