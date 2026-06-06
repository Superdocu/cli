// Package output renders API responses for the terminal.
package output

import (
	"encoding/json"
	"io"
)

// PrintJSON writes body as indented JSON, or verbatim when raw is set or the
// body is not valid JSON.
func PrintJSON(w io.Writer, body []byte, raw bool) error {
	if raw {
		_, err := w.Write(ensureNewline(body))
		return err
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		_, werr := w.Write(ensureNewline(body))
		return werr
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func ensureNewline(b []byte) []byte {
	if len(b) == 0 || b[len(b)-1] == '\n' {
		return b
	}
	return append(b, '\n')
}
