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

package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"chainguard.dev/wolfi-secdb/pkg/types"

	"github.com/spf13/cobra"
)

type Context struct {
	Archs    []string
	Baseurl  string
	Reponame string
	Output   string
	DB       types.Database
}

func (c *Context) Run(args []string) error {
	if c.Reponame == "" {
		return errors.New("repository name not set, use --repo-name")
	}

	c.DB.Apkurl = "{{urlprefix}}/{{reponame}}/{{arch}}/{{pkg.name}}-{{pkg.ver}}.apk"
	c.DB.Archs = c.Archs
	c.DB.Urlprefix = c.Baseurl
	c.DB.Reponame = c.Reponame

	for _, dir := range args {
		if err := c.ProcessDir(dir); err != nil {
			return err
		}
	}

	buf, err := json.MarshalIndent(c.DB, "", "  ")
	if err != nil {
		return err
	}

	outputDir := filepath.Dir(c.Output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(c.Output, buf, 0644); err != nil {
		return err
	}

	return nil
}

func (c *Context) ProcessDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("walking directory %s: %w", dir, err)
	}

	for _, file := range files {
		mp, err := types.LoadMelangePackage(filepath.Join(dir, file.Name()))
		if err != nil {
			return fmt.Errorf("loading %s: %w", file.Name(), err)
		}

		c.DB.Packages = append(c.DB.Packages, mp.Entry())
	}

	return nil
}

type Option func(*Context) error

func WithArchs(archs []string) Option {
	return func(c *Context) error {
		c.Archs = archs
		return nil
	}
}

func WithReponame(repoName string) Option {
	return func(c *Context) error {
		c.Reponame = repoName
		return nil
	}
}

func WithBaseurl(baseUrl string) Option {
	return func(c *Context) error {
		c.Baseurl = baseUrl
		return nil
	}
}

func WithOutput(output string) Option {
	return func(c *Context) error {
		c.Output = output
		return nil
	}
}

func NewContext(opts ...Option) (*Context, error) {
	ctx := &Context{
		Baseurl: "https://packages.wolfi.dev",
		Output:  "security.json",
	}

	for _, opt := range opts {
		if err := opt(ctx); err != nil {
			return nil, err
		}
	}

	return ctx, nil
}

func Generate() *cobra.Command {
	var repoName string
	var baseUrl string
	var output string
	var archs []string

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate a security database",
		Long:    "Generate a security database.",
		Example: "  wolfi-secdb generate ./repo ...",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options := []Option{
				WithArchs(archs),
				WithReponame(repoName),
				WithBaseurl(baseUrl),
				WithOutput(output),
			}

			gc, err := NewContext(options...)
			if err != nil {
				return err
			}

			return gc.Run(args)
		},
	}

	cmd.Flags().StringVar(&repoName, "repo-name", "", "the repository name to use")
	cmd.Flags().StringVar(&baseUrl, "base-url", "https://packages.wolfi.dev", "the repository base URL to use")
	cmd.Flags().StringVar(&output, "output-filename", "security.json", "the output filename to use")
	cmd.Flags().StringSliceVar(&archs, "archs", []string{"x86_64"}, "the package architectures the security database is for")

	return cmd
}
