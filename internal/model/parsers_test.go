package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseScontrolOutput_Nodes(t *testing.T) {
	data := readTestData(t, "nodes.txt")
	entries := parseScontrolOutput(data)

	require.NotEmpty(t, entries, "should parse node entries")

	node := entries[0]
	assert.Len(t, entries, 8888, "unexpected node count")
	assert.Len(t, node, 32, "unexpected field count")

	assert.Equal(t, "linux1", node["NodeName"])
	assert.Equal(t, "ALLOCATED", node["State"])
	assert.Equal(t, "64", node["CPUTot"])

	node = entries[501]
	assert.Contains(t, node["Partitions"], "mathematics")
}

func TestParseScontrolOutput_Jobs(t *testing.T) {
	data := readTestData(t, "jobs.txt")
	entries := parseScontrolOutput(data)

	require.NotEmpty(t, entries, "should parse job entries")

	job := entries[0]
	assert.Len(t, entries, 630)
	assert.Len(t, job, 60, "unexpected field count")

	assert.Equal(t, "6833", job["JobId"], "JobId should match")
	assert.Equal(t, "RUNNING", job["JobState"], "JobState should match")
	assert.Equal(t, "general", job["Partition"], "Partition should match")
}

func TestParseScontrolOutput_Partitions(t *testing.T) {
	data := readTestData(t, "partitions.txt")
	entries := parseScontrolOutput(data)
	require.NotEmpty(t, entries, "should parse partition entries")

	partition := entries[0]
	assert.Len(t, entries, 7)
	assert.Len(t, partition, 36, "unexpected field count")

	assert.Equal(t, "general", entries[0]["PartitionName"], "first partition should match")
	assert.Equal(t, "chemistry", entries[1]["PartitionName"], "second partition should match")
}

func TestParseScontrolOutput_EmptyInput(t *testing.T) {
	entries := parseScontrolOutput("")
	assert.Empty(t, entries, "empty input should return empty entries")
}

func TestParseScontrolOutput_InvalidLines(t *testing.T) {
	input := "Header line\nInvalid line\nKey=Value\n"
	entries := parseScontrolOutput(input)

	require.Len(t, entries, 1, "should only parse valid key=value lines")
	assert.Equal(t, "Value", entries[0]["Key"])
}

func TestSafeGetFromMap(t *testing.T) {
	testMap := map[string]string{
		"exists": "value",
	}

	assert.Equal(t, "value", safeGetFromMap(testMap, "exists"))
	assert.Empty(t, safeGetFromMap(testMap, "missing"))
}

// TestParseSacctOutput_WithSacctDelimiter_AIGenerated verifies parseSacctOutput correctly handles
// the triple-pipe delimiter used by sacct commands.
// AI-generated test for delimiter parsing functionality.
func TestParseSacctOutput_WithSacctDelimiter_AIGenerated(t *testing.T) {
	input := "JobID|||User|||State|||Partition\n123|||john|||RUNNING|||general\n456|||jane|||COMPLETED|||compute"
	entries := parseSacctOutput(input, SACCT_DELIMITER)

	require.Len(t, entries, 2, "should parse 2 entries")

	assert.Equal(t, "123", entries[0]["JobID"])
	assert.Equal(t, "john", entries[0]["User"])
	assert.Equal(t, "RUNNING", entries[0]["State"])
	assert.Equal(t, "general", entries[0]["Partition"])

	assert.Equal(t, "456", entries[1]["JobID"])
	assert.Equal(t, "jane", entries[1]["User"])
	assert.Equal(t, "COMPLETED", entries[1]["State"])
	assert.Equal(t, "compute", entries[1]["Partition"])
}

// TestParseSacctOutput_WithSacctmgrDelimiter_AIGenerated verifies parseSacctOutput correctly handles
// the single-pipe delimiter used by sacctmgr commands.
// AI-generated test for delimiter parsing functionality.
func TestParseSacctOutput_WithSacctmgrDelimiter_AIGenerated(t *testing.T) {
	input := "Account|Descr|Org\nroot|default root account|root\nstudent|student|local student\ntest|testumgebung|local staff"
	entries := parseSacctOutput(input, SACCTMGR_DELIMITER)

	require.Len(t, entries, 3, "should parse 3 entries")

	assert.Equal(t, "root", entries[0]["Account"])
	assert.Equal(t, "default root account", entries[0]["Descr"])
	assert.Equal(t, "root", entries[0]["Org"])

	assert.Equal(t, "student", entries[1]["Account"])
	assert.Equal(t, "student", entries[1]["Descr"])
	assert.Equal(t, "local student", entries[1]["Org"])

	assert.Equal(t, "test", entries[2]["Account"])
	assert.Equal(t, "testumgebung", entries[2]["Descr"])
	assert.Equal(t, "local staff", entries[2]["Org"])
}

// TestParseSacctOutput_EmptyInput_AIGenerated verifies parseSacctOutput handles empty input correctly.
// AI-generated test for edge case handling.
func TestParseSacctOutput_EmptyInput_AIGenerated(t *testing.T) {
	entries := parseSacctOutput("", SACCT_DELIMITER)
	assert.Empty(t, entries, "empty input should return empty entries")
}

// TestParseSacctOutput_OnlyHeader_AIGenerated verifies parseSacctOutput handles header-only input correctly.
// AI-generated test for edge case handling.
func TestParseSacctOutput_OnlyHeader_AIGenerated(t *testing.T) {
	entries := parseSacctOutput("Header1|||Header2|||Header3", SACCT_DELIMITER)
	assert.Empty(t, entries, "only header should return empty entries")
}

// TestParseSacctOutput_MismatchedFieldCount_AIGenerated verifies parseSacctOutput correctly skips
// rows with mismatched field counts.
// AI-generated test for error handling.
func TestParseSacctOutput_MismatchedFieldCount_AIGenerated(t *testing.T) {
	input := "Col1|Col2|Col3\nvalue1|value2|value3\nvalue4|value5"
	entries := parseSacctOutput(input, SACCTMGR_DELIMITER)

	require.Len(t, entries, 1, "should skip rows with mismatched field count")
	assert.Equal(t, "value1", entries[0]["Col1"])
	assert.Equal(t, "value2", entries[0]["Col2"])
	assert.Equal(t, "value3", entries[0]["Col3"])
}

// readTestData helper reads test data from testdata directory
func readTestData(t *testing.T, filename string) string {
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	require.NoError(t, err, "failed to read test data file")
	return string(data)
}
