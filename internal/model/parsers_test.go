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
	entries := parseScontrolOutput("NodeName=", data)

	require.NotEmpty(t, entries, "should parse node entries")

	node := entries[0]
	assert.Len(t, entries, 888)
	assert.Len(t, node, 32, "unexpected field count")

	assert.Equal(t, "linux1", node["NodeName"])
	assert.Equal(t, "MIXED", node["State"])
	assert.Equal(t, "64", node["CPUTot"])

	node = entries[501]
	assert.Contains(t, node["Partitions"], "mathematics")
}

func TestParseScontrolOutput_Jobs(t *testing.T) {
	data := readTestData(t, "jobs.txt")
	entries := parseScontrolOutput("JobId=", data)

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
	entries := parseScontrolOutput("PartitionName=", data)
	require.NotEmpty(t, entries, "should parse partition entries")

	partition := entries[0]
	assert.Len(t, entries, 7)
	assert.Len(t, partition, 36, "unexpected field count")

	assert.Equal(t, "general", entries[0]["PartitionName"], "first partition should match")
	assert.Equal(t, "chemistry", entries[1]["PartitionName"], "second partition should match")
}

func TestParseScontrolOutput_EmptyInput(t *testing.T) {
	entries := parseScontrolOutput("", "")
	assert.Empty(t, entries, "empty input should return empty entries")
}

func TestParseScontrolOutput_InvalidLines(t *testing.T) {
	input := "Header line\nInvalid line\nKey=Value\n"
	entries := parseScontrolOutput("", input)

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

// readTestData helper reads test data from testdata directory
func readTestData(t *testing.T, filename string) string {
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	require.NoError(t, err, "failed to read test data file")
	return string(data)
}
