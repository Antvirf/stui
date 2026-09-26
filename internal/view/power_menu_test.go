package view

import (
	"testing"

	"github.com/rivo/tview"
)

func TestNodePowerCommand(t *testing.T) {
	command := nodePowerCommand(
		map[string]bool{"node-b": true, "node-a": true},
		"POWER_UP",
	)

	want := `scontrol update NodeName="node-a,node-b" State="POWER_UP"`
	if command != want {
		t.Errorf("nodePowerCommand() = %q, want %q", command, want)
	}
}

func TestNodePowerStateIndex(t *testing.T) {
	tests := []struct {
		shortcut rune
		want     int
	}{
		{'u', 0},
		{'U', 0},
		{'d', 1},
		{'D', 1},
		{'a', 2},
		{'A', 2},
		{'f', 3},
		{'F', 3},
		{'x', -1},
	}

	for _, tt := range tests {
		if got := nodePowerStateIndex(tt.shortcut); got != tt.want {
			t.Errorf("nodePowerStateIndex(%q) = %d, want %d", tt.shortcut, got, tt.want)
		}
	}
}

func TestNodePowerOptions(t *testing.T) {
	states := []string{
		"POWER_UP",
		"POWER_DOWN",
		"POWER_DOWN_ASAP",
		"POWER_DOWN_FORCE",
	}

	if len(nodePowerOptions) != len(states) {
		t.Fatalf("nodePowerOptions has %d entries, want %d", len(nodePowerOptions), len(states))
	}
	for index, state := range states {
		option := nodePowerOptions[index]
		if option.state != state {
			t.Errorf("option %d state = %q, want %q", index, option.state, state)
		}
		if option.description == "" {
			t.Errorf("option %q has no description", state)
		}
	}
}

func TestShowNodePowerMenuFocusesMenu(t *testing.T) {
	app := &App{
		App:   tview.NewApplication(),
		Pages: tview.NewPages(),
	}

	app.ShowNodePowerMenu(map[string]bool{"node-a": true})

	if _, ok := app.App.GetFocus().(*tview.List); !ok {
		t.Errorf("menu focus = %T, want *tview.List", app.App.GetFocus())
	}
}
