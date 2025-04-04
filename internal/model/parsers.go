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

// parseScontrolOutput parses the scontrol show node output into a slice of maps
func parseScontrolOutput(prefix string, output string) []map[string]string {
	var nodes []map[string]string
	currentNode := make(map[string]string)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a new node entry
		if strings.HasPrefix(line, prefix) {
			if len(currentNode) > 0 {
				nodes = append(nodes, currentNode)
			}
			currentNode = make(map[string]string)
		}

		// Parse key=value pairs
		pairs := strings.Fields(line)
		for _, pair := range pairs {
			if idx := strings.Index(pair, "="); idx > 0 {
				key := pair[:idx]
				value := pair[idx+1:]
				currentNode[key] = value
			}
		}
	}

	// Add the last node if it exists
	if len(currentNode) > 0 {
		nodes = append(nodes, currentNode)
	}

	return nodes
}
