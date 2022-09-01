# wolfi-secdb

Tool for generating Wolfi security databases

## Usage

To create a security database for a given project, you
can do something like:

```shell
$ wolfi-secdb generate ./path/to/source-repo \
   --base-url https://packages.wolfi.dev/... \
   --output-filename security/your-repo-name.json \
   --repo-name your-repo-name
```

For the Wolfi distribution, there is a GitHub action
located in [chainguard-dev/actions][gha].

   [gha]: https://github.com/chainguard-dev/actions

## Specification

Wolfi security databases are based on Alpine's
security database format, presenting a serialized
JSON graph.

### Root

The root of the graph has these fields:

- `urlprefix`: The prefix for all URLs.  In Wolfi itself,
  this is `https://packages.wolfi.dev`.

- `apkurl`: The pattern used to deduce the package URL.  In Wolfi itself,
  this is `{{urlprefix}}/{{reponame}}/{{arch}}/{{pkg.name}}-{{pkg.ver}}.apk`

- `reponame`: The name of the repository, such as `bootstrap/stage3`.

- `archs`: The architectures for packages built in the repository.
  In Wolfi itself, this is presently `[ "x86_64" ]`.

- `packages`: A list of package objects which have security updates.

### Package entries

A package object is a JSON object which has a single `pkg` object
underneath it, which has the following fields:

- `name`: The name of the package.

- `secfixes`: An object containing version identifiers and lists of
  well-known vulnerability identifiers fixed by the package version.

### Example

```json
{
  "urlprefix": "https://packages.wolfi.dev",
  "apkurl": "{{urlprefix}}/{{reponame}}/{{arch}}/{{pkg.name}}-{{pkg.ver}}.apk",
  "reponame": "example/repo",
  "archs": ["x86_64"],
  "packages": [
    {
      "pkg": {
        "name": "foo",
        "secfixes": {
          "1.2.3-r1": [
            "CVE-9999-99999"
          ]
        }
      }
    }
  ]
}
```
