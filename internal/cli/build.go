package cli

import (
	"github.com/Superdocu/cli/internal/spec"
	"github.com/spf13/cobra"
)

// addAPICommands mounts one parent command per group and one subcommand per
// operation underneath it.
func addAPICommands(root *cobra.Command, s *spec.Spec) {
	groups := make(map[string]*cobra.Command, len(s.Groups))
	for _, name := range s.Groups {
		gc := &cobra.Command{
			Use:   name,
			Short: "Operations on " + name,
		}
		groups[name] = gc
		root.AddCommand(gc)
	}
	for i := range s.Operations {
		op := s.Operations[i]
		if gc := groups[op.Group]; gc != nil {
			gc.AddCommand(newOpCmd(op))
		}
	}
}
