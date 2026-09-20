package model

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/antvirf/stui/internal/logger"
)

// safeGetFromMap retrieves a value from a map by key, returning an empty string if the key does not exist
func safeGetFromMap(input map[string]string, key string) string {
	value, exists := input[key]
	if exists {
		return value
	}
	return ""
}

// parseScontrolOutput parses the scontrol show output into a slice of maps
func parseScontrolOutput(output string) (entries []map[string]string) {
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

// parseScontrolJobsOutput parses the scontrol show job output into a slice of maps
func parseScontrolJobsOutput(output string) (jobs []map[string]string) {
	for _, job := range strings.Split(output, "\n\n") {
		currentJob := make(map[string]string)

		// Trim surrounding whitespace and ignore empty lines
		job = strings.TrimSpace(job)
		if job == "" {
			continue
		}

		// At this point, we have a multiline block of data for a particular job
		// We split them by newline, which gives us a slice of strings that are
		// of two types:
		// Multiple fields in the string, e.g.: RunTime=00:00:30 TimeLimit=UNLIMITED TimeMin=N/A
		// Single field in the string, e.g.: Comment=this is a multi word comment of job-general-5
		//
		// Single field lines are special cases; multiple field lines are the general case.

		for _, line := range strings.Split(job, "\n") {
			// Trim surrounding whitespace and ignore empty lines
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Which type of line is this - check for special cases
			isSingleFieldLine := false
			for _, prefix := range SCONTROL_JOB_SINGLE_LINE_COLUMNS {
				if strings.HasPrefix(line, prefix) {
					isSingleFieldLine = true
					break
				}
			}

			// Special case handling: First line of a job with ID, Array info, and Job name
			// This needs special handling as JobName may have spaces, for example:
			// JobId=11111 ArrayJobId=888888 ArrayTaskId=1 JobName=Name with spaces in it
			// Our check for this line uses hardcoded keys, hence the code block does too.
			if strings.Contains(line, "JobId=") && strings.Contains(line, "JobName=") {
				jobNameIndex := strings.Index(line, " JobName=")
				currentJob["JobName"] = line[jobNameIndex+9:]

				// The data before JobName can be processed normally, so we let the usual logic handle it
				// and do *not* do `continue` here.
				line = line[:jobNameIndex]
			}

			// Special case handling: Fields on a single line
			if isSingleFieldLine {
				// Split by the first =
				if idx := strings.Index(line, "="); idx > 0 {
					key := line[:idx]
					value := line[idx+1:]
					currentJob[key] = strings.TrimSpace(value)
					continue
				}
			}

			// Otherwise, deal with as normal
			// Parse key=value pairs into current entry
			pairs := strings.Fields(line)
			for _, pair := range pairs {
				if idx := strings.Index(pair, "="); idx > 0 {
					key := pair[:idx]
					value := pair[idx+1:]

					currentJob[key] = value
				}
			}
		}
		// Only add entries that contain at least 1 key=value pair
		if len(currentJob) != 0 {
			jobs = append(jobs, currentJob)
		}
	}

	return jobs
}

// parseSacctOutput parses the sacct/sacctmgr output into a slice of maps
func parseSacctOutput(output string, delimiter string) (entries []map[string]string) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return entries // Return empty if there are no rows or only a header
	}

	// Parse the header
	header := strings.Split(lines[0], delimiter)

	// Parse the data rows
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}

		// Split the line into fields
		fields := strings.Split(line, delimiter)
		if len(fields) != len(header) {
			continue // Skip rows that don't match the header length, if we get some random garbage
		}

		// Create a map for the current entry
		currentEntry := make(map[string]string)
		for i, key := range header {
			currentEntry[key] = fields[i]
		}

		entries = append(entries, currentEntry)
	}

	return entries
}

// parseSacctMgrRunawayJobsOutput parses the sacctmgr runaway jobs format into a slice of maps
func parseSacctMgrRunawayJobsOutput(output string) (entries []map[string]string) {
	if os.Getenv("STUI_TESTING") != "" {
		logger.Debugf("STUI_TESTING env var set, using hardcoded runaway jobs data...")
		rawOut, _ := os.ReadFile("./internal/model/testdata/runaway_jobs.txt")
		output = string(rawOut)
	}

	lines := strings.Split(output, "\n")
	headers := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Ignore the note starting line and all empty lines
		if strings.HasPrefix(line, "NOTE:") || line == "" {
			continue
		}

		// If headers length is zero, we haven't set it yet, so let's do so now
		if len(headers) == 0 {
			headers = strings.Split(line, "|")
			continue // Don't process further on this particular run, no data to add
		}

		// Once we reach the start of the dialog lines, stop parsing.
		if strings.HasPrefix(line, "Would you like to fix") {
			break
		}

		// Split the line into fields
		fields := strings.Split(line, "|")
		if len(fields) != len(headers) {
			continue // Skip rows that don't match the header length, if we get some random garbage
		}

		// Create a map for the current entry
		currentEntry := make(map[string]string)
		for i, key := range headers {
			currentEntry[key] = fields[i]
		}

		entries = append(entries, currentEntry)
	}
	return entries
}

// ExpandNodeList expands a Slurm compressed node list notation into individual node names.
// Examples:
//   - "linux1" → "linux1"
//   - "linux[1,5,7]" → "linux1,linux5,linux7"
//   - "linux[1-5]" → "linux1,linux2,linux3,linux4,linux5"
//   - "linux[1,5,7,11-15]" → "linux1,linux5,linux7,linux11,linux12,linux13,linux14,linux15"
//   - "node[01-03]" → "node01,node02,node03" (preserves zero-padding)
//   - "node[1-3],gpu[1-2]" → "node1,node2,node3,gpu1,gpu2" (multiple groups)
func ExpandNodeList(nodeList string) string {
	nodeList = strings.TrimSpace(nodeList)
	if nodeList == "" || nodeList == "N/A" || nodeList == "(null)" {
		return nodeList
	}

	// Handle multiple comma-separated node groups (e.g., "node[1-3],gpu[1-2]")
	// We need to be careful: commas inside brackets are part of the range spec
	var results []string
	for _, group := range splitNodeGroups(nodeList) {
		results = append(results, expandSingleNodeGroup(group)...)
	}

	return strings.Join(results, ",")
}

// splitNodeGroups splits a node list into individual groups, respecting brackets.
// e.g., "node[1,3],gpu[1-2]" → ["node[1,3]", "gpu[1-2]"]
func splitNodeGroups(nodeList string) []string {
	var groups []string
	depth := 0
	start := 0

	for i, ch := range nodeList {
		switch ch {
		case '[':
			depth++
		case ']':
			depth--
		case ',':
			if depth == 0 {
				groups = append(groups, nodeList[start:i])
				start = i + 1
			}
		}
	}
	// Add the last group
	groups = append(groups, nodeList[start:])
	return groups
}

// expandSingleNodeGroup expands a single node group like "linux[1,5,7,11-15]"
func expandSingleNodeGroup(group string) []string {
	group = strings.TrimSpace(group)
	if group == "" {
		return nil
	}

	// Find bracket positions
	bracketStart := strings.Index(group, "[")
	bracketEnd := strings.LastIndex(group, "]")

	// No brackets - it's a simple node name
	if bracketStart == -1 || bracketEnd == -1 || bracketEnd <= bracketStart {
		return []string{group}
	}

	prefix := group[:bracketStart]
	suffix := ""
	if bracketEnd < len(group)-1 {
		suffix = group[bracketEnd+1:]
	}
	rangeSpec := group[bracketStart+1 : bracketEnd]

	// Parse the range specification (comma-separated items, each may be a range)
	var nodes []string
	for _, part := range strings.Split(rangeSpec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Range like "11-15" or "01-05"
			rangeParts := strings.SplitN(part, "-", 2)
			if len(rangeParts) != 2 {
				nodes = append(nodes, prefix+part+suffix)
				continue
			}

			startStr := rangeParts[0]
			endStr := rangeParts[1]

			startNum, err1 := strconv.Atoi(startStr)
			endNum, err2 := strconv.Atoi(endStr)
			if err1 != nil || err2 != nil {
				nodes = append(nodes, prefix+part+suffix)
				continue
			}

			// Determine padding width from the original string
			padWidth := len(startStr)

			for n := startNum; n <= endNum; n++ {
				nodes = append(nodes, prefix+fmt.Sprintf("%0*d", padWidth, n)+suffix)
			}
		} else {
			// Single number like "5" or "05"
			nodes = append(nodes, prefix+part+suffix)
		}
	}

	return nodes
}
