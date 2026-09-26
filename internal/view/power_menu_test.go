package view

import (
	"testing"

	"github.com/rivo/tview"
)

func TestNodePowerCommand(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{"power up", `scontrol power up "node-a,node-b"`},
		{"power down", `scontrol power down "node-a,node-b"`},
		{"power down asap", `scontrol power down asap "node-a,node-b"`},
		{"power down force", `scontrol power down force "node-a,node-b"`},
		{"reboot", `scontrol reboot "node-a,node-b"`},
		{"reboot asap", `scontrol reboot asap "node-a,node-b"`},
		{"cancel_reboot", `scontrol cancel_reboot "node-a,node-b"`},
	}

	for _, tt := range tests {
		if got := nodePowerCommand(map[string]bool{"node-b": true, "node-a": true}, tt.command); got != tt.want {
			t.Errorf("nodePowerCommand(%q) = %q, want %q", tt.command, got, tt.want)
		}
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
		{'r', 4},
		{'R', 4},
		{'b', 5},
		{'B', 5},
		{'c', 6},
		{'C', 6},
		{'x', -1},
	}

	for _, tt := range tests {
		if got := nodePowerStateIndex(tt.shortcut); got != tt.want {
			t.Errorf("nodePowerStateIndex(%q) = %d, want %d", tt.shortcut, got, tt.want)
		}
	}
}

func TestNodePowerOptions(t *testing.T) {
	commands := []string{
		"power up",
		"power down",
		"power down asap",
		"power down force",
		"reboot",
		"reboot asap",
		"cancel_reboot",
	}

	if len(nodePowerOptions) != len(commands) {
		t.Fatalf("nodePowerOptions has %d entries, want %d", len(nodePowerOptions), len(commands))
	}
	for index, command := range commands {
		option := nodePowerOptions[index]
		if option.command != command {
			t.Errorf("option %d command = %q, want %q", index, option.command, command)
		}
		if option.description == "" {
			t.Errorf("option %q has no description", command)
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
