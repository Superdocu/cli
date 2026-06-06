package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Superdocu/cli/internal/client"
	"github.com/Superdocu/cli/internal/output"
	"github.com/Superdocu/cli/internal/spec"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// newOpCmd builds the Cobra command for a single API operation, wiring path
// params as positional args and query/body/multipart fields as flags.
func newOpCmd(op spec.Operation) *cobra.Command {
	cc := &cobra.Command{
		Use:   opUse(op),
		Short: op.Summary,
		Long:  opLong(op),
		Args:  cobra.ExactArgs(len(op.PathParams)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOp(cmd, op, args)
		},
	}
	f := cc.Flags()

	for _, qp := range op.QueryParams {
		if name := flagName(qp.Name); f.Lookup(name) == nil {
			f.String(name, "", stripTicks(qp.Description))
		}
	}

	if op.Body != nil {
		for _, a := range op.Body.Attrs {
			if name := flagName(a.Name); f.Lookup(name) == nil {
				f.String(name, "", attrHelp(a))
			}
		}
		f.String("data-json", "", "Raw JSON:API request body (overrides attribute flags)")
	}

	if mp := op.Multipart; mp != nil {
		for _, tf := range mp.TextFields {
			if name := flagName(tf.Name); f.Lookup(name) == nil {
				f.String(name, "", attrHelp(tf))
			}
		}
		for _, ff := range mp.Files {
			if name := flagName(ff.Name); f.Lookup(name) == nil {
				f.String(name, "", fmt.Sprintf("Path to the file to upload as %q", ff.Name))
			}
		}
		if mp.BareBinary || mp.ArrayBinary {
			f.StringArray("file", nil, "Path to a file to upload (repeatable)")
		}
		f.StringArray("file-field", nil, "Upload a file as field=path (repeatable)")
	}

	return cc
}

func runOp(cmd *cobra.Command, op spec.Operation, args []string) error {
	if api.Token == "" {
		return errors.New("not authenticated — run `superdocu login` or pass --token / SUPERDOCU_TOKEN")
	}
	f := cmd.Flags()

	path := op.Path
	for i, p := range op.PathParams {
		path = strings.ReplaceAll(path, "{"+p.Name+"}", url.PathEscape(args[i]))
	}

	q := url.Values{}
	for _, qp := range op.QueryParams {
		name := flagName(qp.Name)
		if f.Changed(name) {
			v, _ := f.GetString(name)
			q.Set(qp.Name, v)
		}
	}
	for _, kv := range g.queryKV {
		if k, v, ok := strings.Cut(kv, "="); ok {
			q.Set(k, v)
		}
	}

	var body io.Reader
	contentType := ""
	switch {
	case op.Multipart != nil:
		b, ct, err := buildMultipart(f, op)
		if err != nil {
			return err
		}
		body, contentType = b, ct
	case op.Body != nil:
		b, err := buildJSONBody(f, op)
		if err != nil {
			return err
		}
		if b != nil {
			body, contentType = bytes.NewReader(b), "application/json"
		}
	}

	headers := map[string]string{}
	if g.idempotency != "" {
		headers["Idempotency-Key"] = g.idempotency
	}

	resp, err := api.Do(client.Request{
		Method:      op.Method,
		Path:        path,
		Query:       q.Encode(),
		Headers:     headers,
		Body:        body,
		ContentType: contentType,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return render(resp)
}

func buildJSONBody(f *pflag.FlagSet, op spec.Operation) ([]byte, error) {
	if f.Changed("data-json") {
		raw, _ := f.GetString("data-json")
		return []byte(raw), nil
	}
	attrs := map[string]any{}
	for _, a := range op.Body.Attrs {
		name := flagName(a.Name)
		if !f.Changed(name) {
			continue
		}
		v, _ := f.GetString(name)
		attrs[a.Name] = coerce(v, a.Type)
	}
	if len(attrs) == 0 {
		return nil, nil
	}
	data := map[string]any{"attributes": attrs}
	if op.Body.DataType != "" {
		data["type"] = op.Body.DataType
	}
	return json.Marshal(map[string]any{"data": data})
}

func buildMultipart(f *pflag.FlagSet, op spec.Operation) (io.Reader, string, error) {
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	mp := op.Multipart

	addFile := func(field, path string) error {
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()
		part, err := w.CreateFormFile(field, filepath.Base(path))
		if err != nil {
			return err
		}
		_, err = io.Copy(part, src)
		return err
	}

	for _, tf := range mp.TextFields {
		name := flagName(tf.Name)
		if f.Changed(name) {
			v, _ := f.GetString(name)
			if err := w.WriteField(tf.Name, v); err != nil {
				return nil, "", err
			}
		}
	}
	for _, ff := range mp.Files {
		name := flagName(ff.Name)
		if f.Changed(name) {
			p, _ := f.GetString(name)
			if err := addFile(ff.Name, p); err != nil {
				return nil, "", err
			}
		}
	}
	if mp.BareBinary || mp.ArrayBinary {
		field := "file"
		if mp.ArrayBinary {
			field = "files[]"
		}
		paths, _ := f.GetStringArray("file")
		for _, p := range paths {
			if err := addFile(field, p); err != nil {
				return nil, "", err
			}
		}
	}
	if fields, _ := f.GetStringArray("file-field"); len(fields) > 0 {
		for _, kv := range fields {
			field, p, ok := strings.Cut(kv, "=")
			if !ok {
				return nil, "", fmt.Errorf("invalid --file-field %q, expected field=path", kv)
			}
			if err := addFile(field, p); err != nil {
				return nil, "", err
			}
		}
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf, w.FormDataContentType(), nil
}

func render(resp *http.Response) error {
	if g.verbose {
		printResponseHeaders(resp)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" || strings.Contains(ct, "json") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			return client.ParseError(resp, body)
		}
		if len(bytes.TrimSpace(body)) == 0 {
			fmt.Fprintln(os.Stderr, resp.Status)
			return nil
		}
		return output.PrintJSON(os.Stdout, body, g.raw)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return client.ParseError(resp, body)
	}
	out := io.Writer(os.Stdout)
	if g.output != "" {
		fo, err := os.Create(g.output)
		if err != nil {
			return err
		}
		defer fo.Close()
		out = fo
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	if g.output != "" {
		fmt.Fprintln(os.Stderr, "saved to "+g.output)
	}
	return nil
}

func printResponseHeaders(resp *http.Response) {
	fmt.Fprintln(os.Stderr, resp.Proto, resp.Status)
	for _, h := range []string{
		"X-Ratelimit-Limit", "X-Ratelimit-Remaining", "X-Ratelimit-Reset",
		"Idempotent-Replay", "Idempotent-Generated-Key", "Retry-After",
	} {
		if v := resp.Header.Get(h); v != "" {
			fmt.Fprintf(os.Stderr, "%s: %s\n", h, v)
		}
	}
}

func coerce(s, t string) any {
	switch t {
	case "integer":
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
	case "number":
		if fv, err := strconv.ParseFloat(s, 64); err == nil {
			return fv
		}
	case "boolean":
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
	case "array", "object":
		var v any
		if json.Unmarshal([]byte(s), &v) == nil {
			return v
		}
	}
	if len(s) > 0 && (s[0] == '[' || s[0] == '{') {
		var v any
		if json.Unmarshal([]byte(s), &v) == nil {
			return v
		}
	}
	return s
}
