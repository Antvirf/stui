package model

import (
	"strings"
)

// safeGetFromMap
func safeGetFromMap(input map[string]string, key string) string {
	value, exists := input[key]
	if exists {
		return value
	}
	return ""
}

// parseScontrolOutput parses the scontrol show output into a slice of maps
func parseScontrolOutput(prefix string, output string) (entries []map[string]string) {
	for _, line := range strings.Split(output, "\n") {
		// Trim surrounding whitespace and ignore empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse key=value pairs into current entry
		currentEntry := make(map[string]string)
		pairs := strings.Fields(line)
		for i, pair := range pairs {
			if idx := strings.Index(pair, "="); idx > 0 {
				key := pair[:idx]
				value := pair[idx+1:]
				// Special handling for "Reason" key
				// ... as long as it's the last pair on this line
				// ... which we confirm by ensuring there are no more '='
				if key == "Reason" && !strings.Contains(
					strings.Join(pairs[i+1:], " ")[idx+1:],
					"=",
				) {
					// Capture everything after "Reason=" since it's the last key
					// and can contain arbitrary whitespaces and other characters.
					value = strings.Join(pairs[i:], " ")[idx+1:]
				}
				currentEntry[key] = value
			}
		}
		// Only add entries that contain at least 1 key=value pair
		if len(currentEntry) != 0 {
			entries = append(entries, currentEntry)
		}
	}
	return entries
}
