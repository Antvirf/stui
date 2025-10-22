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

// formatMemoryValue converts raw memory values (in bytes) to human-readable format
func formatMemoryValue(raw string) string {
	// Try to parse as integer first, then as float, give up and return original value on failure
	mem, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		if memFloat, err := strconv.ParseFloat(raw, 64); err == nil {
			mem = int64(memFloat)
		} else {
			return raw // Return original if not a number
		}
	}
	return fmt.Sprintf("%.1fG", float64(mem)/1024)
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

				// Format memory-related fields
				if strings.HasSuffix(key, "Mem") ||
					strings.HasSuffix(key, "Memory") {
					value = formatMemoryValue(value)
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

					// Format memory-related fields
					if strings.HasSuffix(key, "Mem") ||
						strings.HasSuffix(key, "Memory") {
						value = formatMemoryValue(value)
					}
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
func parseSacctOutput(output string) (entries []map[string]string) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return entries // Return empty if there are no rows or only a header
	}

	// Parse the header
	header := strings.Split(lines[0], SACCT_DELIMITER)

	// Parse the data rows
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // Skip empty lines
		}

		// Split the line into fields
		fields := strings.Split(line, SACCT_DELIMITER)
		if len(fields) != len(header) {
			continue // Skip rows that don't match the header length, if we get some random garbage
		}

		// Create a map for the current entry
		currentEntry := make(map[string]string)
		for i, key := range header {
			// Format memory-related fields from sacct output too
			if strings.HasSuffix(key, "Mem") || key == "Memory" {
				currentEntry[key] = formatMemoryValue(fields[i])
			} else {
				currentEntry[key] = fields[i]
			}
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
