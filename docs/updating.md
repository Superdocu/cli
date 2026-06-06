# Updating & releasing

Three independent things can change: the **CLI source**, the **embedded API
spec**, and the **published binaries**. Each has its own refresh path.

## Rebuild after a new CLI version

When the repository has new commits (someone improved the CLI):

```sh
cd ~/work/superdocu/cli
git pull
make build      # -> bin/superdocu
# or, to put it on your PATH:
make install    # -> $(go env GOPATH)/bin/superdocu
```

Check what you are running:

```sh
./bin/superdocu --version
```

`make build`/`make install` stamp the version from `git describe`
(`-X main.version=...`), so tagged checkouts report a clean version and untagged
ones report e.g. `v0.2.0-3-gabc1234-dirty`.

Prefer not to build from source? Grab a prebuilt archive for your OS/arch from
the [GitHub releases](https://github.com/Superdocu/cli/releases) and put the
`superdocu` binary on your PATH.

## Refresh the embedded API spec

The whole command tree is generated from `openapi/api.yaml`, embedded at build
time. When the API changes (new endpoint, new field, renamed filter), re-pull
the spec and rebuild — no Go code to touch:

```sh
make gen        # curl https://developers.superdocu.com/api-docs/v2/api.yaml -> openapi/api.yaml
make build
```

Review and validate the change before committing:

```sh
git diff --stat openapi/api.yaml      # what moved in the spec
go test ./...                         # naming/parsing still holds (no command collisions)
./bin/superdocu --help                # eyeball the new/changed commands
git add openapi/api.yaml && git commit -m "Refresh embedded API v2 spec"
```

`make gen` points at `SPEC_URL` (overridable), defaulting to the published spec:
`https://developers.superdocu.com/api-docs/v2/api.yaml`.

## Cut a release

Releases are built by [GoReleaser](https://goreleaser.com) from a tag
(`.goreleaser.yaml` builds linux/darwin/windows × amd64/arm64).

```sh
# 1. make sure main is green and the spec is current
go test ./... && make build

# 2. tag with semver
git tag -a v0.3.0 -m "v0.3.0"
git push origin v0.3.0

# 3. build + publish the GitHub release (needs GITHUB_TOKEN with repo scope)
goreleaser release --clean
```

Dry-run the build without publishing:

```sh
goreleaser release --snapshot --clean   # artifacts land in ./dist, no upload
```

Install GoReleaser if needed: `brew install goreleaser`.
