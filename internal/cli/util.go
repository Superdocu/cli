package cli

import (
	"fmt"
	"strings"

	"github.com/Superdocu/cli/internal/spec"
)

// flagName maps an API parameter or attribute name to a CLI flag name, e.g.
// "filter[status]" -> "filter-status", "per_page" -> "per-page".
func flagName(apiName string) string {
	return spec.Kebab(apiName)
}

// stripTicks removes backticks so pflag does not treat a quoted word in a
// description as the flag's value-type placeholder.
func stripTicks(s string) string {
	return strings.ReplaceAll(s, "`", "")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// opUse renders the command usage line, e.g. "get <id> <wid>".
func opUse(op spec.Operation) string {
	parts := []string{op.Name}
	for _, p := range op.PathParams {
		parts = append(parts, "<"+spec.Kebab(p.Name)+">")
	}
	return strings.Join(parts, " ")
}

// opLong builds the long help: summary, description, method+path.
func opLong(op spec.Operation) string {
	var b strings.Builder
	if op.Summary != "" {
		b.WriteString(op.Summary)
		b.WriteString("\n\n")
	}
	if op.Description != "" {
		b.WriteString(op.Description)
		b.WriteString("\n\n")
	}
	fmt.Fprintf(&b, "%s %s", strings.ToUpper(op.Method), op.Path)
	return b.String()
}

func attrHelp(a spec.BodyAttr) string {
	help := a.Description
	if help == "" {
		help = a.Name
	}
	var notes []string
	if a.Required {
		notes = append(notes, "required")
	}
	if a.Type != "" && a.Type != "string" {
		notes = append(notes, a.Type)
	}
	if len(a.Enum) > 0 {
		notes = append(notes, "one of: "+strings.Join(a.Enum, ", "))
	}
	if len(notes) > 0 {
		help += " (" + strings.Join(notes, "; ") + ")"
	}
	return stripTicks(help)
}
