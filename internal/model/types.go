package model

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/logger"
)

// CellValue represents a typed cell value with display and programmatic access.
// Each column maintains a consistent type - all cells in a column are instances
// of the same concrete type (e.g., all IntegerValue, all MemoryValue, etc.).
// Individual cells can be null/invalid via the IsNull() flag.
type CellValue interface {
	// Display returns the user-facing formatted string
	Display() string

	// Raw returns the underlying value for programmatic use
	// Returns nil if IsNull() is true
	Raw() interface{}

	// Compare returns -1, 0, or 1 for sorting (handles type-appropriate comparison)
	// Null values sort to the end
	Compare(other CellValue) int

	// Type returns the value type
	Type() CellType

	// IsNull returns true if this value is null/invalid/N/A
	IsNull() bool
}

// CellType represents the type of data stored in a cell
type CellType int

const (
	TypeString CellType = iota
	TypeInteger
	TypeFloat
	TypeDuration
	TypeMemory
	TypeTimestamp
	TypeRatio
)

func (ct CellType) String() string {
	switch ct {
	case TypeString:
		return "String"
	case TypeInteger:
		return "Integer"
	case TypeFloat:
		return "Float"
	case TypeDuration:
		return "Duration"
	case TypeMemory:
		return "Memory"
	case TypeTimestamp:
		return "Timestamp"
	case TypeRatio:
		return "Ratio"
	default:
		return "Unknown"
	}
}

// isNullIndicator checks if a string represents a null/missing value
func isNullIndicator(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == "N/A" || s == "(null)"
}

// compareNulls handles null comparison logic used by all Compare() methods.
// Returns (comparisonResult, shouldReturn) where shouldReturn indicates if the comparison is complete.
// Null values always sort to the end of lists.
func compareNulls(thisNull, otherNull bool) (int, bool) {
	if thisNull && otherNull {
		return 0, true // Both null, equal
	}
	if thisNull {
		return 1, true // This is null, sorts after non-null
	}
	if otherNull {
		return -1, true // Other is null, this sorts before it
	}
	return 0, false // Neither null, continue with type-specific comparison
}

// StringValue represents a string value
type StringValue struct {
	value  string
	isNull bool
}

// NewStringValue creates a new StringValue
// Note: Null indicators like "N/A" and "(null)" are stored as-is internally,
// but Display() always normalizes them to "N/A" for consistent user-facing output.
// This preserves the original input while providing UI consistency.
func NewStringValue(s string) *StringValue {
	// For strings, we generally don't treat empty as null
	// But we do treat explicit null indicators as null
	if s == "N/A" || s == "(null)" {
		return &StringValue{value: s, isNull: true}
	}
	return &StringValue{value: s, isNull: false}
}

func (v *StringValue) Display() string {
	if v.isNull {
		return "N/A"
	}
	return v.value
}

func (v *StringValue) Raw() interface{} {
	if v.isNull {
		return nil
	}
	return v.value
}

func (v *StringValue) Compare(other CellValue) int {
	otherStr, ok := other.(*StringValue)
	if !ok {
		// Fallback to string comparison for different types
		return strings.Compare(v.Display(), other.Display())
	}

	// Handle null values
	if cmp, done := compareNulls(v.isNull, otherStr.isNull); done {
		return cmp
	}

	return strings.Compare(v.value, otherStr.value)
}

func (v *StringValue) Type() CellType {
	return TypeString
}

func (v *StringValue) IsNull() bool {
	return v.isNull
}

// IntegerValue represents an integer value
type IntegerValue struct {
	value   int64
	isNull  bool
	rawText string // Store original for display if needed
}

// NewIntegerValue creates a new IntegerValue
func NewIntegerValue(s string) *IntegerValue {
	// Check for null indicators
	if isNullIndicator(s) {
		return &IntegerValue{value: 0, isNull: true, rawText: s}
	}

	s = strings.TrimSpace(s)

	// Try to parse
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		logger.Debugf("Failed to parse integer value '%s', treating as null", s)
		return &IntegerValue{value: 0, isNull: true, rawText: s}
	}

	return &IntegerValue{value: val, isNull: false, rawText: s}
}

func (v *IntegerValue) Display() string {
	if v.isNull {
		return "N/A"
	}
	// Use original text if available to preserve formatting
	if v.rawText != "" {
		return v.rawText
	}
	return strconv.FormatInt(v.value, 10)
}

func (v *IntegerValue) Raw() interface{} {
	if v.isNull {
		return nil
	}
	return v.value
}

func (v *IntegerValue) Compare(other CellValue) int {
	otherInt, ok := other.(*IntegerValue)
	if !ok {
		return strings.Compare(v.Display(), other.Display())
	}

	// Handle null values
	if cmp, done := compareNulls(v.isNull, otherInt.isNull); done {
		return cmp
	}

	// Numeric comparison
	if v.value < otherInt.value {
		return -1
	} else if v.value > otherInt.value {
		return 1
	}
	return 0
}

func (v *IntegerValue) Type() CellType {
	return TypeInteger
}

func (v *IntegerValue) IsNull() bool {
	return v.isNull
}

// FloatValue represents a floating-point value
type FloatValue struct {
	value   float64
	isNull  bool
	rawText string
}

// NewFloatValue creates a new FloatValue
func NewFloatValue(s string) *FloatValue {
	// Check for null indicators
	if isNullIndicator(s) {
		return &FloatValue{value: 0.0, isNull: true, rawText: s}
	}

	s = strings.TrimSpace(s)

	// Try to parse
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Debugf("Failed to parse float value '%s', treating as null", s)
		return &FloatValue{value: 0.0, isNull: true, rawText: s}
	}

	return &FloatValue{value: val, isNull: false, rawText: s}
}

func (v *FloatValue) Display() string {
	if v.isNull {
		return "N/A"
	}
	// Use original text to preserve formatting
	if v.rawText != "" {
		return v.rawText
	}
	return strconv.FormatFloat(v.value, 'f', -1, 64)
}

func (v *FloatValue) Raw() interface{} {
	if v.isNull {
		return nil
	}
	return v.value
}

func (v *FloatValue) Compare(other CellValue) int {
	otherFloat, ok := other.(*FloatValue)
	if !ok {
		return strings.Compare(v.Display(), other.Display())
	}

	// Handle null values
	if cmp, done := compareNulls(v.isNull, otherFloat.isNull); done {
		return cmp
	}

	// Numeric comparison
	if v.value < otherFloat.value {
		return -1
	} else if v.value > otherFloat.value {
		return 1
	}
	return 0
}

func (v *FloatValue) Type() CellType {
	return TypeFloat
}

func (v *FloatValue) IsNull() bool {
	return v.isNull
}

// MemoryValue represents a memory value (stored as MB, displayed as GB)
type MemoryValue struct {
	megabytes int64 // Stored in MB, displayed as GB
	isNull    bool
}

// NewMemoryValue creates a new MemoryValue
func NewMemoryValue(s string) *MemoryValue {
	// Check for null indicators
	if isNullIndicator(s) {
		return &MemoryValue{megabytes: 0, isNull: true}
	}

	s = strings.TrimSpace(s)

	// Parse memory string to megabytes
	megabytes := parseMemoryString(s)

	// Treat as null if parse likely failed (returned 0 but input doesn't look like a zero value)
	if megabytes == 0 && !looksLikeZero(s) {
		logger.Debugf("Failed to parse memory value '%s', treating as null", s)
		return &MemoryValue{megabytes: 0, isNull: true}
	}

	return &MemoryValue{megabytes: megabytes, isNull: false}
}

// looksLikeZero checks if a string appears to be a zero value (not a parse failure)
func looksLikeZero(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	// Strip sacct suffixes
	s = strings.TrimSuffix(s, "N")
	s = strings.TrimSuffix(s, "C")

	// Check common zero patterns
	return s == "0" || s == "0M" || s == "0G" || s == "0K" || s == "0T" ||
		s == "0.0" || s == "0.0M" || s == "0.0G" || s == "0.0K" || s == "0.0T" ||
		s == "0MB" || s == "0GB" || s == "0KB" || s == "0TB"
}

func (v *MemoryValue) Display() string {
	if v.isNull {
		return "N/A"
	}
	// Always display as GB with 1 decimal place (megabytes / 1024 = gigabytes)
	return fmt.Sprintf("%.1fG", float64(v.megabytes)/1024)
}

func (v *MemoryValue) Raw() interface{} {
	if v.isNull {
		return nil
	}
	return v.megabytes
}

func (v *MemoryValue) Compare(other CellValue) int {
	otherMem, ok := other.(*MemoryValue)
	if !ok {
		return strings.Compare(v.Display(), other.Display())
	}

	// Handle null values
	if cmp, done := compareNulls(v.isNull, otherMem.isNull); done {
		return cmp
	}

	// Megabyte comparison
	if v.megabytes < otherMem.megabytes {
		return -1
	} else if v.megabytes > otherMem.megabytes {
		return 1
	}
	return 0
}

func (v *MemoryValue) Type() CellType {
	return TypeMemory
}

func (v *MemoryValue) IsNull() bool {
	return v.isNull
}

// parseMemoryString converts memory string to megabytes (MB)
// Handles: "2.5G", "1024M", "16000" (raw MB), etc.
// Returns value in MB (base unit for internal storage)
func parseMemoryString(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Try parsing as raw megabytes (MB) first
	if mb, err := strconv.ParseInt(s, 10, 64); err == nil {
		return mb
	}

	// Try parsing as float (for decimal values)
	if mb, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(mb)
	}

	// Parse with unit suffix
	s = strings.ToUpper(s)

	// Strip sacct-specific suffixes (n=per-node, c=per-core)
	// Example: "43Gn" → "43G", "100Mc" → "100M"
	s = strings.TrimSuffix(s, "N")
	s = strings.TrimSuffix(s, "C")

	// Extract number and unit
	var numStr string
	var unit string
	for i, ch := range s {
		if (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' {
			numStr += string(ch)
		} else {
			unit = s[i:]
			break
		}
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	// Convert based on unit - base unit is MB
	// Display formula: display_GB = stored_MB / 1024
	switch strings.TrimSpace(unit) {
	case "K", "KB":
		return int64(num / 1024) // KB to MB
	case "M", "MB":
		return int64(num) // Already in MB
	case "G", "GB":
		return int64(num * 1024) // GB to MB
	case "T", "TB":
		return int64(num * 1024 * 1024) // TB to MB
	default:
		// No unit, assume it's already in MB (the raw values from scontrol are in MB)
		return int64(num)
	}
}

// DurationValue represents a time duration
type DurationValue struct {
	duration    time.Duration
	isNull      bool
	rawText     string
	isUnlimited bool
	isInfinite  bool
}

// NewDurationValue creates a new DurationValue
func NewDurationValue(s string) *DurationValue {
	s = strings.TrimSpace(s)
	sUpper := strings.ToUpper(s)

	// Handle special values
	switch sUpper {
	case "UNLIMITED":
		return &DurationValue{
			duration:    time.Duration(math.MaxInt64),
			isNull:      false,
			rawText:     s,
			isUnlimited: true,
		}
	case "INFINITE":
		return &DurationValue{
			duration:   time.Duration(math.MaxInt64),
			isNull:     false,
			rawText:    s,
			isInfinite: true,
		}
	case "NOT_SET", "", "N/A", "(null)":
		return &DurationValue{
			duration: 0,
			isNull:   true,
			rawText:  s,
		}
	}

	// Parse normal duration
	duration := parseSlurmDuration(s)
	if duration == 0 && s != "0" && s != "00:00:00" && s != "0:00:00" {
		// Parse failed
		logger.Debugf("Failed to parse duration value '%s', treating as null", s)
		return &DurationValue{
			duration: 0,
			isNull:   true,
			rawText:  s,
		}
	}

	return &DurationValue{
		duration: duration,
		isNull:   false,
		rawText:  s,
	}
}

func (v *DurationValue) Display() string {
	if v.isNull {
		return "N/A"
	}
	// Display original format from Slurm
	return v.rawText
}

func (v *DurationValue) Raw() interface{} {
	if v.isNull {
		return nil
	}
	return v.duration
}

func (v *DurationValue) Compare(other CellValue) int {
	otherDur, ok := other.(*DurationValue)
	if !ok {
		return strings.Compare(v.Display(), other.Display())
	}

	// Handle null values
	if cmp, done := compareNulls(v.isNull, otherDur.isNull); done {
		return cmp
	}

	// Special case: both unlimited or infinite
	if (v.isUnlimited || v.isInfinite) && (otherDur.isUnlimited || otherDur.isInfinite) {
		return 0
	}

	// Duration comparison
	if v.duration < otherDur.duration {
		return -1
	} else if v.duration > otherDur.duration {
		return 1
	}
	return 0
}

func (v *DurationValue) Type() CellType {
	return TypeDuration
}

func (v *DurationValue) IsNull() bool {
	return v.isNull
}

// parseSlurmDuration parses Slurm duration formats into time.Duration
// Formats: "HH:MM:SS", "D-HH:MM:SS", "MM:SS", etc.
func parseSlurmDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	var days, hours, minutes, seconds int64

	// Check for day component (e.g., "5-12:30:00")
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) == 2 {
			days, _ = strconv.ParseInt(parts[0], 10, 64)
			s = parts[1]
		}
	}

	// Parse time components (HH:MM:SS or MM:SS)
	timeParts := strings.Split(s, ":")
	switch len(timeParts) {
	case 3: // HH:MM:SS
		hours, _ = strconv.ParseInt(timeParts[0], 10, 64)
		minutes, _ = strconv.ParseInt(timeParts[1], 10, 64)
		seconds, _ = strconv.ParseInt(timeParts[2], 10, 64)
	case 2: // MM:SS
		minutes, _ = strconv.ParseInt(timeParts[0], 10, 64)
		seconds, _ = strconv.ParseInt(timeParts[1], 10, 64)
	case 1: // SS
		seconds, _ = strconv.ParseInt(timeParts[0], 10, 64)
	}

	totalSeconds := days*86400 + hours*3600 + minutes*60 + seconds
	return time.Duration(totalSeconds) * time.Second
}

// TimestampValue represents an absolute point in time (distinct from duration which is a time span)
// Handles Slurm timestamp formats like "2025-10-22T19:05:58" (ISO 8601)
// Also handles special values: "Unknown", "None", "N/A"
type TimestampValue struct {
	timestamp time.Time
	isNull    bool
	rawText   string
}

// NewTimestampValue creates a new TimestampValue
func NewTimestampValue(s string) *TimestampValue {
	// Check for null indicators
	if isNullIndicator(s) {
		return &TimestampValue{isNull: true, rawText: s}
	}

	s = strings.TrimSpace(s)

	// Try parsing ISO 8601 format used by Slurm
	formats := []string{
		"2006-01-02T15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return &TimestampValue{timestamp: t, isNull: false, rawText: s}
		}
	}

	// Parse failed - treat as null
	logger.Debugf("Failed to parse timestamp value '%s', treating as null", s)
	return &TimestampValue{isNull: true, rawText: s}
}

func (v *TimestampValue) Display() string {
	if v.isNull {
		return "N/A"
	}
	// Show original format from Slurm
	return v.rawText
}

func (v *TimestampValue) Raw() interface{} {
	if v.isNull {
		return nil
	}
	return v.timestamp
}

func (v *TimestampValue) Compare(other CellValue) int {
	otherTs, ok := other.(*TimestampValue)
	if !ok {
		return strings.Compare(v.Display(), other.Display())
	}

	// Handle null values
	if cmp, done := compareNulls(v.isNull, otherTs.isNull); done {
		return cmp
	}

	// Time comparison
	if v.timestamp.Before(otherTs.timestamp) {
		return -1
	} else if v.timestamp.After(otherTs.timestamp) {
		return 1
	}
	return 0
}

func (v *TimestampValue) Type() CellType {
	return TypeTimestamp
}

func (v *TimestampValue) IsNull() bool {
	return v.isNull
}

// RatioValue represents a computed ratio between two numeric values
// Display shows: "numerator / denominator (percentage%)"
// Sorts by: the ratio value (0.0 to 1.0+)
type RatioValue struct {
	numerator   CellValue // The first value (e.g., AllocMem)
	denominator CellValue // The second value (e.g., RealMemory)
	ratio       float64   // Computed ratio (numerator/denominator)
	isNull      bool      // True if either value is null or denominator is zero
}

// NewRatioValue creates a new RatioValue from two numeric values
func NewRatioValue(numerator, denominator CellValue) *RatioValue {
	// Check if either is null
	if numerator.IsNull() || denominator.IsNull() {
		return &RatioValue{
			numerator:   numerator,
			denominator: denominator,
			ratio:       0,
			isNull:      true,
		}
	}

	// Extract raw numeric values
	numRaw := numerator.Raw()
	denRaw := denominator.Raw()

	// Convert to float64 for ratio calculation
	var numFloat, denFloat float64

	switch v := numRaw.(type) {
	case int64:
		numFloat = float64(v)
	case float64:
		numFloat = v
	default:
		// Non-numeric type, cannot compute ratio
		logger.Debugf("Cannot compute ratio: numerator is non-numeric type %T", numRaw)
		return &RatioValue{numerator: numerator, denominator: denominator, isNull: true}
	}

	switch v := denRaw.(type) {
	case int64:
		denFloat = float64(v)
	case float64:
		denFloat = v
	default:
		// Non-numeric type, cannot compute ratio
		logger.Debugf("Cannot compute ratio: denominator is non-numeric type %T", denRaw)
		return &RatioValue{numerator: numerator, denominator: denominator, isNull: true}
	}

	// Handle division by zero
	if denFloat == 0 {
		return &RatioValue{
			numerator:   numerator,
			denominator: denominator,
			ratio:       0,
			isNull:      true,
		}
	}

	ratio := numFloat / denFloat
	return &RatioValue{
		numerator:   numerator,
		denominator: denominator,
		ratio:       ratio,
		isNull:      false,
	}
}

func (v *RatioValue) Display() string {
	if v.isNull {
		return "N/A"
	}

	// Show: "numerator / denominator (percentage%)"
	percentage := v.ratio * 100
	return fmt.Sprintf("%s / %s (%.0f%%)",
		v.numerator.Display(),
		v.denominator.Display(),
		percentage)
}

func (v *RatioValue) Raw() interface{} {
	if v.isNull {
		return nil
	}
	return v.ratio // Return the ratio for sorting/programmatic use
}

func (v *RatioValue) Compare(other CellValue) int {
	otherRatio, ok := other.(*RatioValue)
	if !ok {
		return strings.Compare(v.Display(), other.Display())
	}

	// Handle null values
	if cmp, done := compareNulls(v.isNull, otherRatio.isNull); done {
		return cmp
	}

	// Compare by ratio
	if v.ratio < otherRatio.ratio {
		return -1
	} else if v.ratio > otherRatio.ratio {
		return 1
	}
	return 0
}

func (v *RatioValue) Type() CellType {
	return TypeRatio
}

func (v *RatioValue) IsNull() bool {
	return v.isNull
}

// parseTypedValue converts a raw string to the appropriate CellValue type
func parseTypedValue(rawValue string, columnType CellType) CellValue {
	switch columnType {
	case TypeInteger:
		return NewIntegerValue(rawValue)
	case TypeFloat:
		return NewFloatValue(rawValue)
	case TypeMemory:
		return NewMemoryValue(rawValue)
	case TypeDuration:
		return NewDurationValue(rawValue)
	case TypeTimestamp:
		return NewTimestampValue(rawValue)
	case TypeRatio:
		// Ratio values are created directly, not parsed from strings
		logger.Debugf("TypeRatio should not be parsed from string, using StringValue")
		return NewStringValue(rawValue)
	default:
		return NewStringValue(rawValue)
	}
}
