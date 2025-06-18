package view

import "github.com/gdamore/tcell/v2"

var (
	BAD_STATE_COLOR     = tcell.ColorRed
	MID_STATE_COLOR     = tcell.ColorOrange
	SUCCESS_STATE_COLOR = tcell.ColorLightGreen

	// This map will get processed first, and takes priority
	STATE_COLORS_MAP_HIGH_PRIORITY = map[string]tcell.Color{
		// Job States
		"BOOT_FAIL":     BAD_STATE_COLOR,
		"CANCELLED":     BAD_STATE_COLOR,
		"DEADLINE":      BAD_STATE_COLOR,
		"FAILED":        BAD_STATE_COLOR,
		"NODE_FAIL":     BAD_STATE_COLOR,
		"OUT_OF_MEMORY": BAD_STATE_COLOR,
		"PREEMPTED":     BAD_STATE_COLOR,
		"TIMEOUT":       BAD_STATE_COLOR,

		// Node states
		"DOWN":         BAD_STATE_COLOR,
		"POWER_DOWN":   BAD_STATE_COLOR,
		"POWERED_DOWN": BAD_STATE_COLOR,
		"NO_RESPOND":   BAD_STATE_COLOR,
		"DRAINED":      BAD_STATE_COLOR,
		"UNKNOWN":      BAD_STATE_COLOR,
		"UNK":          BAD_STATE_COLOR,
		"BLOCKED":      BAD_STATE_COLOR,
		"FAIL":         BAD_STATE_COLOR,

		// Shared states
		"COMPLETED": SUCCESS_STATE_COLOR,
	}

	STATE_COLORS_MAP = map[string]tcell.Color{
		// Job states
		"PENDING":   MID_STATE_COLOR,
		"SUSPENDED": MID_STATE_COLOR,

		// Node states
		"DRAINING":      MID_STATE_COLOR,
		"POWERING_DOWN": MID_STATE_COLOR,
		"DRAIN":         MID_STATE_COLOR,
		"MAINT":         MID_STATE_COLOR,
		"FUTURE":        MID_STATE_COLOR,
	}
)
