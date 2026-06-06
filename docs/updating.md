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

Releases are **automated**: pushing a `v*` tag triggers
`.github/workflows/release.yml`, which runs [GoReleaser](https://goreleaser.com)
to build the binaries (linux/darwin/windows × amd64/arm64), checksums and the
Homebrew cask, then publishes the GitHub Release and the cask.

```sh
go test ./...                      # main should be green (CI runs this too)
git tag -a v0.3.0 -m "v0.3.0"
git push origin v0.3.0             # -> Actions builds & publishes the release
```

That's it — no local GoReleaser needed.

### One-time prerequisites for the Homebrew cask

The cask is published to a **separate tap repo**, which the default Actions
`GITHUB_TOKEN` cannot write to. Before the first release:

1. Create the repo `Superdocu/homebrew-tap` (can be empty).
2. Create a Personal Access Token with `repo` scope on that repo.
3. Add it to this repo as the `TAP_GITHUB_TOKEN` Actions secret
   (Settings → Secrets and variables → Actions).

Without it, the release still publishes binaries; only the cask step fails. Once
set, users can `brew install Superdocu/tap/superdocu`. The binaries are not
Apple-notarized, so the cask strips the macOS quarantine flag on install.

### Validate or release locally (fallback)

```sh
brew install goreleaser
goreleaser check                          # validate .goreleaser.yaml
goreleaser release --snapshot --clean     # build everything into ./dist, no upload
goreleaser release --clean                # real release; needs GITHUB_TOKEN (+ TAP_GITHUB_TOKEN)
```

## Continuous integration

`.github/workflows/ci.yml` runs `go vet`, `go test` and `go build` on every push
to `main` and on pull requests.
