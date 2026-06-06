// Package superdocu embeds the OpenAPI document that drives the CLI, kept at the
// module root so the spec stays discoverable (openapi/api.yaml) while the binary
// entrypoint lives under cmd/superdocu (so `go install .../cmd/superdocu` names
// it `superdocu`).
package superdocu

import _ "embed"

//go:embed openapi/api.yaml
var SpecData []byte
