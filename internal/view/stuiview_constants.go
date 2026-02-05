package view

import (
	"regexp"

	"github.com/gdamore/tcell/v2"
)

type StatePattern struct {
	State  string
	Color  tcell.Color
	Regexp *regexp.Regexp
}

var (
	BAD_STATE_COLOR      = tcell.ColorRed
	MID_STATE_COLOR      = tcell.ColorOrange
	SUCCESS_STATE_COLOR  = tcell.ColorSpringGreen
	RESUMING_STATE_COLOR = tcell.ColorLightCyan
	INACTIVE_STATE_COLOR = tcell.Color248 // Light Gray

	// StatePriorityList defines the order of precedence for color-coding states.
	// Earlier entries in the list take priority over later ones.
	StatePriorityList = []struct {
		State string
		Color tcell.Color
	}{
		// 1. Critical/Bad States (Red)
		{"BOOT_FAIL", BAD_STATE_COLOR},
		{"CANCELLED", BAD_STATE_COLOR},
		{"DEADLINE", BAD_STATE_COLOR},
		{"FAILED", BAD_STATE_COLOR},
		{"NODE_FAIL", BAD_STATE_COLOR},
		{"OUT_OF_MEMORY", BAD_STATE_COLOR},
		{"PREEMPTED", BAD_STATE_COLOR},
		{"TIMEOUT", BAD_STATE_COLOR},
		{"DOWN", BAD_STATE_COLOR},
		{"POWER_DOWN", BAD_STATE_COLOR},
		{"NO_RESPOND", BAD_STATE_COLOR},
		{"NOT_RESPONDING", BAD_STATE_COLOR},
		{"DRAINED", BAD_STATE_COLOR},
		{"UNKNOWN", BAD_STATE_COLOR},
		{"UNK", BAD_STATE_COLOR},
		{"BLOCKED", BAD_STATE_COLOR},
		{"FAIL", BAD_STATE_COLOR},

		// 2. Warning/Transition States (Orange)
		{"SUSPENDED", MID_STATE_COLOR},
		{"PENDING", MID_STATE_COLOR},
		{"DRAINING", MID_STATE_COLOR},
		{"POWERING_DOWN", MID_STATE_COLOR},
		{"DRAIN", MID_STATE_COLOR},
		{"MAINT", MID_STATE_COLOR},

		// 3. Resuming/Active Transition States (Cyan)
		{"POWERING_UP", RESUMING_STATE_COLOR},
		{"RESUMING", RESUMING_STATE_COLOR},
		{"REBOOTING", RESUMING_STATE_COLOR},

		// 4. Inactive/Future States (Gray)
		{"POWERED_DOWN", INACTIVE_STATE_COLOR},
		{"FUTURE", INACTIVE_STATE_COLOR},

		// 5. Success States (Green)
		{"COMPLETED", SUCCESS_STATE_COLOR},
	}

	StatePatterns []StatePattern
)

func init() {
	for _, entry := range StatePriorityList {
		StatePatterns = append(StatePatterns, StatePattern{
			State:  entry.State,
			Color:  entry.Color,
			Regexp: regexp.MustCompile(`\b` + regexp.QuoteMeta(entry.State) + `\b`),
		})
	}
}
