package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGeneratedCredentialPlumbingAndCustomAuthCommandsRemainWired(t *testing.T) {
	cmd, err := NewRootCommand()
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range [][]string{
		{"configure"},
		{"login"},
		{"logout"},
		{"env"},
		{"env", "install"},
		{"env", "uninstall"},
		{"exec"},
		{"doctor"},
		{"auth"},
		{"auth", "login"},
		{"auth", "env"},
		{"auth", "env", "install"},
		{"auth", "whoami"},
		{"auth", "logout"},
		{"keys"},
		{"keys", "create-saved"},
	} {
		requireCommandPath(t, cmd, path...)
	}

	rootLogin := requireCommandPath(t, cmd, "login")
	authLogin := requireCommandPath(t, cmd, "auth", "login")
	for _, loginCmd := range []*cobra.Command{rootLogin, authLogin} {
		for _, flag := range []string{"callback-port", "install-env", "print-env", "no-store"} {
			if loginCmd.Flags().Lookup(flag) == nil {
				t.Fatalf("%s is missing custom login flag --%s", loginCmd.CommandPath(), flag)
			}
		}
	}

	if requireCommandPath(t, cmd, "keys", "create-saved").Flags().Lookup("include-byok-in-limit") == nil {
		t.Fatal("keys create-saved is missing the custom --include-byok-in-limit flag")
	}
}

func requireCommandPath(t *testing.T, root *cobra.Command, path ...string) *cobra.Command {
	t.Helper()
	cmd := root
	for _, name := range path {
		var next *cobra.Command
		for _, child := range cmd.Commands() {
			if child.Name() == name || commandHasAlias(child, name) {
				next = child
				break
			}
		}
		if next == nil {
			t.Fatalf("%s is missing child command %q; available: %s", cmd.CommandPath(), name, childCommandNames(cmd))
		}
		cmd = next
	}
	return cmd
}

func commandHasAlias(cmd *cobra.Command, name string) bool {
	for _, alias := range cmd.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}

func childCommandNames(cmd *cobra.Command) string {
	names := make([]string, 0, len(cmd.Commands()))
	for _, child := range cmd.Commands() {
		names = append(names, child.Name())
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
