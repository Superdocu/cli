# Superdocu CLI

Command-line client for the [Superdocu API v2](https://developers.superdocu.com/api-v2/tutorial).

The entire command tree is **generated at runtime from the embedded OpenAPI
spec** (`openapi/api.yaml`). Every endpoint is exposed automatically, so
refreshing the spec (`make gen`) keeps the CLI in sync with the API — no
per-endpoint code to maintain.

## Install

```sh
make build      # -> bin/superdocu
make install    # -> $GOBIN/superdocu
```

Or grab a binary from the GitHub releases (built via GoReleaser).

## Authenticate

Tokens are issued from the Superdocu admin UI (Bearer JWT). They are stored in
the OS keychain.

```sh
superdocu login                 # prompts for the token, validates via /me
superdocu login --token <jwt>   # non-interactive
echo "<jwt>" | superdocu login  # from a pipe / CI
superdocu whoami                # show current identity, permissions, expiry
superdocu logout
```

Host and token can also come from `--host` / `--token` flags or the
`SUPERDOCU_HOST` / `SUPERDOCU_TOKEN` environment variables.

## Usage

Commands are grouped by resource; run any of them with `--help`:

```sh
superdocu --help
superdocu contacts --help
superdocu contacts list --filter-status active --per-page 50
superdocu contacts get <id>
superdocu contacts create --first-name Alice --last-name Durand --email alice@example.com
superdocu contacts invite <id>
superdocu documents-groups approve <id>
superdocu documents-groups request <id> --requested-message "Merci d'ajouter le recto"
superdocu step-requests reject <id> --rejection-message "Document illisible"
superdocu imports import --file contacts.csv
superdocu shared-documents create <id> --name "Contrat" --file contrat.pdf
superdocu contacts export --filter-status active -o contacts.csv
superdocu dashboard get
```

### Conventions

- **Path parameters** are positional arguments, in path order.
- **Query parameters** (incl. `filter[...]`, pagination, `sort`) are flags;
  `filter[status]` becomes `--filter-status`. Anything not modelled can be
  passed with `--query key=value` (repeatable).
- **Request bodies** (JSON:API) are flat `--attribute` flags, auto-wrapped into
  `{"data":{"type":...,"attributes":{...}}}`. For complex payloads use
  `--data-json '<raw json>'`.
- **File uploads** (multipart) use `--file <path>` (and `--<field>` for named
  fields, or `--file-field name=path` to override the field name).
- **Idempotency**: pass `--idempotency-key <uuid>` on mutations.
- **Output**: pretty JSON by default, `--raw` for compact, `-o <file>` to save
  binary/CSV streams (e.g. exports, document downloads).
- `-v/--verbose` prints rate-limit and idempotency response headers to stderr.

## Keep in sync with the API

```sh
make gen        # re-download openapi/api.yaml from developers.superdocu.com
make build      # rebuild; new/changed endpoints appear automatically
```
