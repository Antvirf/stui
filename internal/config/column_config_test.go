package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseColumnConfigLine_RatioColumn(t *testing.T) {
	input := "Field1,Field2%%Field3,Field4"
	cols, err := parseColumnConfigLine(input)
	require.NoError(t, err)
	require.Equal(t, 3, len(*cols))

	// Check first column
	assert.Equal(t, "Field1", (*cols)[0].RawName)
	assert.False(t, (*cols)[0].ComputedColumn)

	// Check ratio column
	assert.Equal(t, "Field2//Field3", (*cols)[1].RawName) // Should be normalized to //
	assert.Equal(t, "Field2 / Field3", (*cols)[1].DisplayName)
	assert.True(t, (*cols)[1].ComputedColumn)
	assert.Equal(t, "ratio", (*cols)[1].ComputationType)
	assert.True(t, (*cols)[1].DividedByColumn)

	// Check third column
	assert.Equal(t, "Field4", (*cols)[2].RawName)
	assert.False(t, (*cols)[2].ComputedColumn)
}

func TestGetColumnFields_WithRatioColumn(t *testing.T) {
	input := "Field1,Field2%%Field3,Field4"
	cols, err := parseColumnConfigLine(input)
	require.NoError(t, err)

	fields := GetColumnFields(cols)

	// Should extract all individual fields (ratio column is split into components)
	assert.Equal(t, 4, len(fields))
	assert.Equal(t, "Field1", fields[0])
	assert.Equal(t, "Field2", fields[1])
	assert.Equal(t, "Field3", fields[2])
	assert.Equal(t, "Field4", fields[3])

	// Should NOT contain the %% operator
	for _, field := range fields {
		assert.NotContains(t, field, "%%")
	}
}

func TestGetColumnFields_MixedOperators(t *testing.T) {
	input := "A,B//C,D%%E,F++"
	cols, err := parseColumnConfigLine(input)
	require.NoError(t, err)

	fields := GetColumnFields(cols)

	// Should extract: A, B, C, D, E, F (++ is stripped)
	assert.Equal(t, 6, len(fields))
	assert.Contains(t, fields, "A")
	assert.Contains(t, fields, "B")
	assert.Contains(t, fields, "C")
	assert.Contains(t, fields, "D")
	assert.Contains(t, fields, "E")
	assert.Contains(t, fields, "F")
}

func TestParseColumnConfigLine_DividedColumn(t *testing.T) {
	input := "Field1//Field2"
	cols, err := parseColumnConfigLine(input)
	require.NoError(t, err)
	require.Equal(t, 1, len(*cols))

	assert.Equal(t, "Field1//Field2", (*cols)[0].RawName)
	assert.Equal(t, "Field1/Field2", (*cols)[0].DisplayName)
	assert.True(t, (*cols)[0].DividedByColumn)
	assert.False(t, (*cols)[0].ComputedColumn)
}

func TestParseColumnConfigLine_FullWidthColumn(t *testing.T) {
	input := "Field1++"
	cols, err := parseColumnConfigLine(input)
	require.NoError(t, err)
	require.Equal(t, 1, len(*cols))

	assert.Equal(t, "Field1++", (*cols)[0].RawName)
	assert.Equal(t, "Field1", (*cols)[0].DisplayName)
	assert.True(t, (*cols)[0].FullWidthColumn)
}
