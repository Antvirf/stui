package model

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var nodeListRegex = regexp.MustCompile(`([^[]+?)\[(.+?)\]`)

// ExpandNodeList expands Slurm node list notation (e.g., "linux[1,5,7,11-15]")
// into a space-separated list of fully expanded node names.
// This makes array job node lists searchable via regex.
func ExpandNodeList(nodeList string) string {
	if !strings.Contains(nodeList, "[") {
		return strings.ReplaceAll(nodeList, ",", " ")
	}

	var expanded []string
	matches := nodeListRegex.FindAllStringSubmatch(nodeList, -1)
	if matches == nil {
		return strings.ReplaceAll(nodeList, ",", " ")
	}

	for _, match := range matches {
		expanded = append(expanded, expandBracketed(match[1], match[2])...)
	}

	sort.Strings(expanded)
	return strings.Join(expanded, " ")
}

func expandBracketed(prefix string, indices string) []string {
	var nodes []string
	for _, part := range strings.Split(indices, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			if len(rangeParts) == 2 {
				if start, err1 := strconv.Atoi(rangeParts[0]); err1 == nil {
					if end, err2 := strconv.Atoi(rangeParts[1]); err2 == nil {
						for i := start; i <= end; i++ {
							nodes = append(nodes, prefix+strconv.Itoa(i))
						}
					}
				}
			}
		} else {
			nodes = append(nodes, prefix+part)
		}
	}
	return nodes
}
