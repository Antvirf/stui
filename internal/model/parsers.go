package model

import "strings"

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
		for _, pair := range pairs {
			if idx := strings.Index(pair, "="); idx > 0 {
				key := pair[:idx]
				value := pair[idx+1:]
				currentEntry[key] = value
			}
		}
		// Only add entires that contain at least 1 key=value pair
		// e.g. if you have no jobs, scontrol returns a single line with "No jobs in the system"
		if len(currentEntry) != 0 {
			entries = append(entries, currentEntry)
		}
	}
	return entries
}
