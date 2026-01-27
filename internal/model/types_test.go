package model

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== StringValue Tests ====================

func TestStringValue_Valid(t *testing.T) {
	v := NewStringValue("hello")
	assert.False(t, v.IsNull())
	assert.Equal(t, "hello", v.Display())
	assert.Equal(t, "hello", v.Raw())
	assert.Equal(t, TypeString, v.Type())
}

func TestStringValue_Null(t *testing.T) {
	v := NewStringValue("N/A")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
	assert.Nil(t, v.Raw())
}

func TestStringValue_Empty(t *testing.T) {
	v := NewStringValue("")
	assert.False(t, v.IsNull()) // Empty strings are valid for StringValue
	assert.Equal(t, "", v.Display())
	assert.Equal(t, "", v.Raw())
}

func TestStringValue_Compare(t *testing.T) {
	a := NewStringValue("apple")
	b := NewStringValue("banana")
	n := NewStringValue("N/A")

	assert.Equal(t, -1, a.Compare(b)) // apple < banana
	assert.Equal(t, 1, b.Compare(a))  // banana > apple
	assert.Equal(t, 0, a.Compare(NewStringValue("apple")))

	// Null sorting
	assert.Equal(t, 1, n.Compare(a))  // null sorts after
	assert.Equal(t, -1, a.Compare(n)) // non-null sorts before
	assert.Equal(t, 0, n.Compare(NewStringValue("N/A")))
}

// ==================== IntegerValue Tests ====================

func TestIntegerValue_Valid(t *testing.T) {
	v := NewIntegerValue("42")
	assert.False(t, v.IsNull())
	assert.Equal(t, "42", v.Display())
	assert.Equal(t, int64(42), v.Raw())
	assert.Equal(t, TypeInteger, v.Type())
}

func TestIntegerValue_Zero(t *testing.T) {
	v := NewIntegerValue("0")
	assert.False(t, v.IsNull())
	assert.Equal(t, "0", v.Display())
	assert.Equal(t, int64(0), v.Raw())
}

func TestIntegerValue_Negative(t *testing.T) {
	v := NewIntegerValue("-100")
	assert.False(t, v.IsNull())
	assert.Equal(t, "-100", v.Display())
	assert.Equal(t, int64(-100), v.Raw())
}

func TestIntegerValue_Invalid(t *testing.T) {
	v := NewIntegerValue("not-a-number")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
	assert.Nil(t, v.Raw())
}

func TestIntegerValue_Empty(t *testing.T) {
	v := NewIntegerValue("")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
}

func TestIntegerValue_ExplicitNull(t *testing.T) {
	v := NewIntegerValue("N/A")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
	assert.Nil(t, v.Raw())
}

func TestIntegerValue_Compare(t *testing.T) {
	v1 := NewIntegerValue("1")
	v2 := NewIntegerValue("2")
	v10 := NewIntegerValue("10")
	v20 := NewIntegerValue("20")
	vNull := NewIntegerValue("N/A")

	// Numeric sorting (not lexical!)
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, -1, v2.Compare(v10))
	assert.Equal(t, -1, v10.Compare(v20))
	assert.Equal(t, 1, v20.Compare(v10))
	assert.Equal(t, 1, v10.Compare(v2))

	// Null sorting
	assert.Equal(t, 1, vNull.Compare(v1))
	assert.Equal(t, -1, v1.Compare(vNull))
	assert.Equal(t, 0, vNull.Compare(NewIntegerValue("")))
}

// ==================== FloatValue Tests ====================

func TestFloatValue_Valid(t *testing.T) {
	v := NewFloatValue("3.14")
	assert.False(t, v.IsNull())
	assert.Equal(t, "3.14", v.Display())
	assert.Equal(t, float64(3.14), v.Raw())
	assert.Equal(t, TypeFloat, v.Type())
}

func TestFloatValue_Zero(t *testing.T) {
	v := NewFloatValue("0.0")
	assert.False(t, v.IsNull())
	assert.Equal(t, float64(0.0), v.Raw())
}

func TestFloatValue_Negative(t *testing.T) {
	v := NewFloatValue("-2.5")
	assert.False(t, v.IsNull())
	assert.Equal(t, float64(-2.5), v.Raw())
}

func TestFloatValue_Invalid(t *testing.T) {
	v := NewFloatValue("abc")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
	assert.Nil(t, v.Raw())
}

func TestFloatValue_Empty(t *testing.T) {
	v := NewFloatValue("")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
}

func TestFloatValue_Compare(t *testing.T) {
	v1 := NewFloatValue("1.5")
	v2 := NewFloatValue("2.7")
	v10 := NewFloatValue("10.1")
	vNull := NewFloatValue("N/A")

	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, -1, v2.Compare(v10))
	assert.Equal(t, 1, v10.Compare(v2))

	// Null sorting
	assert.Equal(t, 1, vNull.Compare(v1))
	assert.Equal(t, -1, v1.Compare(vNull))
}

// ==================== MemoryValue Tests ====================

func TestMemoryValue_Bytes(t *testing.T) {
	v := NewMemoryValue("1024")
	assert.False(t, v.IsNull())
	assert.Equal(t, "1.0G", v.Display()) // 1024 MB = 1 GB in display
	assert.Equal(t, int64(1024), v.Raw())
}

func TestMemoryValue_WithUnit_M(t *testing.T) {
	v := NewMemoryValue("512M")
	assert.False(t, v.IsNull())
	assert.Equal(t, int64(512), v.Raw()) // 512 MB stored as 512
}

func TestMemoryValue_WithUnit_G(t *testing.T) {
	v := NewMemoryValue("2.5G")
	assert.False(t, v.IsNull())
	// 2.5G = 2.5 * 1024 = 2560 MB
	assert.Equal(t, int64(2560), v.Raw())
}

func TestMemoryValue_Zero(t *testing.T) {
	v := NewMemoryValue("0")
	assert.False(t, v.IsNull())
	assert.Equal(t, "0.0G", v.Display())
	assert.Equal(t, int64(0), v.Raw())
}

func TestMemoryValue_Invalid(t *testing.T) {
	v := NewMemoryValue("invalid")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
	assert.Nil(t, v.Raw())
}

func TestMemoryValue_Empty(t *testing.T) {
	v := NewMemoryValue("")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
}

func TestMemoryValue_Compare(t *testing.T) {
	v512M := NewMemoryValue("512M")
	v1G := NewMemoryValue("1G")
	v2G := NewMemoryValue("2G")
	vNull := NewMemoryValue("N/A")

	assert.Equal(t, -1, v512M.Compare(v1G))
	assert.Equal(t, -1, v1G.Compare(v2G))
	assert.Equal(t, 1, v2G.Compare(v1G))

	// Null sorting
	assert.Equal(t, 1, vNull.Compare(v1G))
	assert.Equal(t, -1, v1G.Compare(vNull))
}

// ==================== DurationValue Tests ====================

func TestDurationValue_HoursMinutesSeconds(t *testing.T) {
	v := NewDurationValue("01:23:45")
	assert.False(t, v.IsNull())
	assert.Equal(t, "01:23:45", v.Display())
	expected := 1*time.Hour + 23*time.Minute + 45*time.Second
	assert.Equal(t, expected, v.Raw())
}

func TestDurationValue_DaysHoursMinutesSeconds(t *testing.T) {
	v := NewDurationValue("5-12:30:00")
	assert.False(t, v.IsNull())
	assert.Equal(t, "5-12:30:00", v.Display())
	expected := 5*24*time.Hour + 12*time.Hour + 30*time.Minute
	assert.Equal(t, expected, v.Raw())
}

func TestDurationValue_MinutesSeconds(t *testing.T) {
	v := NewDurationValue("23:45")
	assert.False(t, v.IsNull())
	assert.Equal(t, "23:45", v.Display())
	expected := 23*time.Minute + 45*time.Second
	assert.Equal(t, expected, v.Raw())
}

func TestDurationValue_Zero(t *testing.T) {
	v := NewDurationValue("00:00:00")
	assert.False(t, v.IsNull())
	assert.Equal(t, time.Duration(0), v.Raw())
}

func TestDurationValue_Unlimited(t *testing.T) {
	v := NewDurationValue("UNLIMITED")
	assert.False(t, v.IsNull())
	assert.Equal(t, "UNLIMITED", v.Display())
	assert.NotNil(t, v.Raw())
}

func TestDurationValue_Infinite(t *testing.T) {
	v := NewDurationValue("INFINITE")
	assert.False(t, v.IsNull())
	assert.Equal(t, "INFINITE", v.Display())
}

func TestDurationValue_NotSet(t *testing.T) {
	v := NewDurationValue("NOT_SET")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
	assert.Nil(t, v.Raw())
}

func TestDurationValue_Empty(t *testing.T) {
	v := NewDurationValue("")
	assert.True(t, v.IsNull())
	assert.Equal(t, "N/A", v.Display())
}

func TestDurationValue_Compare(t *testing.T) {
	v1m := NewDurationValue("00:01:00")
	v5m := NewDurationValue("00:05:00")
	v1h := NewDurationValue("01:00:00")
	vUnlimited := NewDurationValue("UNLIMITED")
	vInfinite := NewDurationValue("INFINITE")
	vNull := NewDurationValue("NOT_SET")

	assert.Equal(t, -1, v1m.Compare(v5m))
	assert.Equal(t, -1, v5m.Compare(v1h))
	assert.Equal(t, 1, v1h.Compare(v5m))

	// Unlimited/Infinite compare equal and sort after regular durations
	assert.Equal(t, 0, vUnlimited.Compare(vInfinite))
	assert.Equal(t, -1, v1h.Compare(vUnlimited))

	// Null sorts to end
	assert.Equal(t, -1, v1h.Compare(vNull))
	assert.Equal(t, 1, vNull.Compare(v1h))
}

// ==================== parseTypedValue Tests ====================

func TestParseTypedValue_Integer(t *testing.T) {
	v := parseTypedValue("42", TypeInteger)
	require.IsType(t, &IntegerValue{}, v)
	assert.Equal(t, int64(42), v.Raw())
}

func TestParseTypedValue_Float(t *testing.T) {
	v := parseTypedValue("3.14", TypeFloat)
	require.IsType(t, &FloatValue{}, v)
	assert.Equal(t, float64(3.14), v.Raw())
}

func TestParseTypedValue_Memory(t *testing.T) {
	v := parseTypedValue("1024M", TypeMemory)
	require.IsType(t, &MemoryValue{}, v)
}

func TestParseTypedValue_Duration(t *testing.T) {
	v := parseTypedValue("01:23:45", TypeDuration)
	require.IsType(t, &DurationValue{}, v)
}

func TestParseTypedValue_String(t *testing.T) {
	v := parseTypedValue("hello", TypeString)
	require.IsType(t, &StringValue{}, v)
	assert.Equal(t, "hello", v.Raw())
}

func TestParseTypedValue_Default(t *testing.T) {
	v := parseTypedValue("anything", CellType(999))
	require.IsType(t, &StringValue{}, v)
}

// ==================== Sorting Integration Tests ====================

func TestSorting_MixedNullsIntegers(t *testing.T) {
	values := []CellValue{
		NewIntegerValue("10"),
		NewIntegerValue("N/A"),
		NewIntegerValue("2"),
		NewIntegerValue(""),
		NewIntegerValue("5"),
	}

	// Sort using Compare
	sort.Slice(values, func(i, j int) bool {
		return values[i].Compare(values[j]) < 0
	})

	// Should be: 2, 5, 10, null, null
	assert.Equal(t, int64(2), values[0].Raw())
	assert.Equal(t, int64(5), values[1].Raw())
	assert.Equal(t, int64(10), values[2].Raw())
	assert.True(t, values[3].IsNull())
	assert.True(t, values[4].IsNull())
}

func TestSorting_IntegersCorrectNumericOrder(t *testing.T) {
	// This is the key test - ensure 1, 10, 2, 20 sorts as 1, 2, 10, 20
	values := []CellValue{
		NewIntegerValue("1"),
		NewIntegerValue("10"),
		NewIntegerValue("2"),
		NewIntegerValue("20"),
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Compare(values[j]) < 0
	})

	// Should be numeric order: 1, 2, 10, 20
	assert.Equal(t, int64(1), values[0].Raw())
	assert.Equal(t, int64(2), values[1].Raw())
	assert.Equal(t, int64(10), values[2].Raw())
	assert.Equal(t, int64(20), values[3].Raw())
}

func TestSorting_MemoryBySize(t *testing.T) {
	values := []CellValue{
		NewMemoryValue("2G"),
		NewMemoryValue("512M"),
		NewMemoryValue("1G"),
		NewMemoryValue("256M"),
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Compare(values[j]) < 0
	})

	// Should be: 256M, 512M, 1G, 2G
	displays := []string{values[0].Display(), values[1].Display(), values[2].Display(), values[3].Display()}
	assert.Equal(t, "0.2G", displays[0]) // 256M
	assert.Equal(t, "0.5G", displays[1]) // 512M
	assert.Equal(t, "1.0G", displays[2]) // 1G
	assert.Equal(t, "2.0G", displays[3]) // 2G
}

func TestSorting_DurationsChronologically(t *testing.T) {
	values := []CellValue{
		NewDurationValue("01:00:00"),
		NewDurationValue("00:05:00"),
		NewDurationValue("UNLIMITED"),
		NewDurationValue("00:30:00"),
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Compare(values[j]) < 0
	})

	// Should be: 5m, 30m, 1h, unlimited
	displays := []string{values[0].Display(), values[1].Display(), values[2].Display(), values[3].Display()}
	assert.Equal(t, "00:05:00", displays[0])
	assert.Equal(t, "00:30:00", displays[1])
	assert.Equal(t, "01:00:00", displays[2])
	assert.Equal(t, "UNLIMITED", displays[3])
}

func TestSorting_StringsLexical(t *testing.T) {
	values := []CellValue{
		NewStringValue("banana"),
		NewStringValue("apple"),
		NewStringValue("N/A"),
		NewStringValue("cherry"),
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i].Compare(values[j]) < 0
	})

	// Should be: apple, banana, cherry, null
	displays := []string{values[0].Display(), values[1].Display(), values[2].Display(), values[3].Display()}
	assert.Equal(t, "apple", displays[0])
	assert.Equal(t, "banana", displays[1])
	assert.Equal(t, "cherry", displays[2])
	assert.Equal(t, "N/A", displays[3])
}

// ==================== GetFieldType Tests ====================

func TestGetFieldType_KnownInteger(t *testing.T) {
	assert.Equal(t, TypeInteger, GetFieldType("CPUTot"))
	assert.Equal(t, TypeInteger, GetFieldType("JobId"))
}

func TestGetFieldType_KnownFloat(t *testing.T) {
	assert.Equal(t, TypeFloat, GetFieldType("CPULoad"))
}

func TestGetFieldType_KnownMemory(t *testing.T) {
	assert.Equal(t, TypeMemory, GetFieldType("AllocMem"))
	assert.Equal(t, TypeMemory, GetFieldType("RealMemory"))
}

func TestGetFieldType_KnownDuration(t *testing.T) {
	assert.Equal(t, TypeDuration, GetFieldType("RunTime"))
	assert.Equal(t, TypeDuration, GetFieldType("Elapsed"))
}

func TestGetFieldType_KnownString(t *testing.T) {
	assert.Equal(t, TypeString, GetFieldType("NodeName"))
	assert.Equal(t, TypeString, GetFieldType("State"))
}

func TestGetFieldType_Unknown(t *testing.T) {
	assert.Equal(t, TypeString, GetFieldType("UnknownField"))
	assert.Equal(t, TypeString, GetFieldType("SomeRandomField"))
}

// Note: Divided columns are always treated as TypeString since they combine multiple values
// (e.g., "0.0G / 15.6G" or "0.00 / 0 / 2")

// ==================== Edge Cases ====================

func TestIntegerValue_LargeNumber(t *testing.T) {
	v := NewIntegerValue("9223372036854775807") // max int64
	assert.False(t, v.IsNull())
	assert.Equal(t, int64(9223372036854775807), v.Raw())
}

func TestMemoryValue_LargeValue(t *testing.T) {
	v := NewMemoryValue("1024G") // 1TB
	assert.False(t, v.IsNull())
	assert.NotNil(t, v.Raw())
}

func TestDurationValue_CaseInsensitive(t *testing.T) {
	v1 := NewDurationValue("unlimited")
	v2 := NewDurationValue("UNLIMITED")
	v3 := NewDurationValue("Unlimited")

	assert.False(t, v1.IsNull())
	assert.False(t, v2.IsNull())
	assert.False(t, v3.IsNull())
	assert.Equal(t, v1.Display(), "unlimited")
	assert.Equal(t, v2.Display(), "UNLIMITED")
}
