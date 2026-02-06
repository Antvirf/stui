package view

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestGetStateColorMapping(t *testing.T) {
	tests := []struct {
		state    string
		expected tcell.Color
		mapped   bool
	}{
		// User provided cases
		{"ALLOCATED-CLOUD", tcell.ColorWhite, false},
		{"ALLOCATED-CLOUD-NOT_RESPONDING-POWERING_UP", RESUMING_STATE_COLOR, true},
		{"ALLOCATED-CLOUD-NOT_RESPONDING-POWERING_UP-POWER_DOWN", RESUMING_STATE_COLOR, true},
		{"ALLOCATED-CLOUD+POWER_DOWN", MID_STATE_COLOR, true},
		{"DOWN", BAD_STATE_COLOR, true},
		{"IDLE-CLOUD", tcell.ColorWhite, false},
		{"IDLE-CLOUD-NOT_RESPONDING-POWERED_DOWN", INACTIVE_STATE_COLOR, true},
		{"IDLE-CLOUD-POWERED_DOWN", INACTIVE_STATE_COLOR, true},
		{"IDLE-CLOUD-POWERING_DOWN", MID_STATE_COLOR, true},

		// Priority edge cases
		{"NOT_RESPONDING", BAD_STATE_COLOR, true},
		{"POWERING_UP", RESUMING_STATE_COLOR, true},
		{"POWERED_DOWN", INACTIVE_STATE_COLOR, true},
		{"DRAINED+POWERED_DOWN", BAD_STATE_COLOR, true},   // DRAINED (Red) > POWERED_DOWN (Gray)
		{"REBOOTING+DRAINED", RESUMING_STATE_COLOR, true}, // REBOOTING (Cyan) > DRAINED (Red)
	}

	for _, tt := range tests {
		color, mapped := GetStateColorMapping(tt.state)
		if mapped != tt.mapped {
			t.Errorf("GetStateColorMapping(%s) mapped = %v, want %v", tt.state, mapped, tt.mapped)
		}
		if mapped && color != tt.expected {
			t.Errorf("GetStateColorMapping(%s) color = %v, want %v", tt.state, color, tt.expected)
		}
	}
}
