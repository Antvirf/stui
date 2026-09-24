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
		prefix string
		want   int
	}{
		{"POWER_U", 0},
		{"POWER_D", 1},
		{"INVALID", -1},
	}

	for _, tt := range tests {
		if got := nodePowerStateIndex(tt.prefix); got != tt.want {
			t.Errorf("nodePowerStateIndex(%q) = %d, want %d", tt.prefix, got, tt.want)
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
