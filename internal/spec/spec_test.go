package spec

import (
	"os"
	"testing"
)

func load(t *testing.T) *Spec {
	t.Helper()
	data, err := os.ReadFile("../../openapi/api.yaml")
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	s, err := Load(data)
	if err != nil {
		t.Fatalf("load spec: %v", err)
	}
	return s
}

func find(t *testing.T, s *Spec, method, path string) *Operation {
	t.Helper()
	for i := range s.Operations {
		if s.Operations[i].Method == method && s.Operations[i].Path == path {
			return &s.Operations[i]
		}
	}
	t.Fatalf("operation %s %s not found", method, path)
	return nil
}

func TestDefaultHost(t *testing.T) {
	if got := load(t).DefaultHost; got != "https://api.superdocu.com" {
		t.Errorf("DefaultHost = %q", got)
	}
}

func TestNoNameCollisionsWithinGroup(t *testing.T) {
	s := load(t)
	seen := map[string]string{}
	for _, o := range s.Operations {
		key := o.Group + "/" + o.Name
		if prev, ok := seen[key]; ok {
			t.Errorf("collision %s: %s and %s %s", key, prev, o.Method, o.Path)
		}
		seen[key] = o.Method + " " + o.Path
	}
}

func TestCommandNames(t *testing.T) {
	s := load(t)
	cases := []struct {
		method, path, group, name string
	}{
		{"get", "/api/v2/contacts", "contacts", "list"},
		{"post", "/api/v2/contacts", "contacts", "create"},
		{"get", "/api/v2/contacts/{id}", "contacts", "get"},
		{"post", "/api/v2/contacts/{id}/invite", "contacts", "invite"},
		{"post", "/api/v2/contacts/tags", "contact-tags", "create"},
		{"get", "/api/v2/documents/{id}/file", "files", "documents-file"},
		{"get", "/api/v2/contacts/{id}/timeline", "timeline", "contacts-timeline-by-id"},
		{"get", "/api/v2/contacts/timeline", "timeline", "contacts-timeline"},
		{"post", "/api/v2/documents_groups/{id}/not_applicable/approve", "documents-groups", "not-applicable-approve"},
		{"patch", "/api/v2/documents_groups/{id}/expiration_date", "documents-groups", "expiration-date"},
		{"post", "/api/v2/contacts/{id}/workflows/{workflow_id}/states/{state_id}/custom_documents_groups", "custom-documents-groups", "create"},
	}
	for _, c := range cases {
		op := find(t, s, c.method, c.path)
		if op.Group != c.group || op.Name != c.name {
			t.Errorf("%s %s => %s/%s, want %s/%s", c.method, c.path, op.Group, op.Name, c.group, c.name)
		}
	}
}

func TestPathParamOrder(t *testing.T) {
	op := find(t, load(t), "get", "/api/v2/contacts/{id}/workflows/{wid}")
	if len(op.PathParams) != 2 || op.PathParams[0].Name != "id" || op.PathParams[1].Name != "wid" {
		t.Errorf("path params = %+v", op.PathParams)
	}
}

func TestJSONBody(t *testing.T) {
	op := find(t, load(t), "post", "/api/v2/contacts")
	if op.Body == nil {
		t.Fatal("expected JSON body")
	}
	if op.Body.DataType != "contact" {
		t.Errorf("DataType = %q", op.Body.DataType)
	}
	if !hasAttr(op.Body.Attrs, "email") {
		t.Errorf("missing email attr: %+v", op.Body.Attrs)
	}
}

func TestMultipartBodies(t *testing.T) {
	s := load(t)

	if mp := find(t, s, "post", "/api/v2/contacts/import").Multipart; mp == nil || !mp.BareBinary {
		t.Errorf("import should be bare-binary multipart: %+v", mp)
	}
	if mp := find(t, s, "post", "/api/v2/documents_groups/{id}/documents").Multipart; mp == nil || !mp.ArrayBinary {
		t.Errorf("documents upload should be array-binary multipart: %+v", mp)
	}
	shared := find(t, s, "post", "/api/v2/contacts/{id}/shared_documents").Multipart
	if shared == nil || len(shared.Files) != 1 || shared.Files[0].Name != "file" {
		t.Errorf("shared_documents should expose a 'file' field: %+v", shared)
	}
	if !hasAttr(shared.TextFields, "name") {
		t.Errorf("shared_documents should expose a 'name' text field: %+v", shared.TextFields)
	}
}

func hasAttr(attrs []BodyAttr, name string) bool {
	for _, a := range attrs {
		if a.Name == name {
			return true
		}
	}
	return false
}
