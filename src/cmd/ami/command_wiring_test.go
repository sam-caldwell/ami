package main

import "testing"

func TestRoot_WiresSubcommands(t *testing.T) {
    c := newRootCmd()
    // Ensure key subcommands are present by command name
    want := []string{"init", "clean", "build", "test", "lint", "mod"}
    for _, name := range want {
        if c.Commands() == nil {
            t.Fatalf("no subcommands registered at root")
        }
        found := false
        for _, sc := range c.Commands() {
            if sc.Name() == name {
                found = true
                break
            }
        }
        if !found {
            t.Fatalf("missing subcommand: %s", name)
        }
    }
}

