package model

// FieldTypeMap maps Slurm field names to their expected types.
// Unknown fields default to TypeString.
var FieldTypeMap = map[string]CellType{
	// Integer fields
	"JobId":       TypeInteger,
	"CPUTot":      TypeInteger,
	"CPUAlloc":    TypeInteger,
	"NumCPUs":     TypeInteger,
	"ReqCPUS":     TypeInteger,
	"AllocCPUS":   TypeInteger,
	"ExitCode":    TypeInteger,
	"NodeCount":   TypeInteger,
	"NumNodes":    TypeInteger,
	"NumTasks":    TypeInteger,
	"Priority":    TypeInteger,
	"ArrayJobId":  TypeInteger,
	"ArrayTaskId": TypeInteger,

	// Float fields
	"CPULoad": TypeFloat,

	// Memory fields
	"AllocMem":   TypeMemory,
	"RealMemory": TypeMemory,
	"ReqMem":     TypeMemory,
	"Memory":     TypeMemory,
	"Mem":        TypeMemory,

	// Duration fields
	"RunTime":    TypeDuration,
	"TimeLimit":  TypeDuration,
	"Elapsed":    TypeDuration,
	"TimeMin":    TypeDuration,
	"SubmitTime": TypeDuration,
	"StartTime":  TypeDuration,
	"EndTime":    TypeDuration,

	// String fields (explicit for documentation)
	"NodeName":       TypeString,
	"Partition":      TypeString,
	"Partitions":     TypeString,
	"State":          TypeString,
	"JobState":       TypeString,
	"Account":        TypeString,
	"User":           TypeString,
	"UserId":         TypeString,
	"JobName":        TypeString,
	"NodeList":       TypeString,
	"QOS":            TypeString,
	"Comment":        TypeString,
	"Reason":         TypeString,
	"Gres":           TypeString,
	"ActiveFeatures": TypeString,
	"CfgTRES":        TypeString,
	"AllocTRES":      TypeString,
	"ReqTRES":        TypeString,
	"SubmitLine":     TypeString,
}

// GetFieldType returns the type for a field name, defaulting to TypeString for unknown fields
func GetFieldType(fieldName string) CellType {
	if cellType, exists := FieldTypeMap[fieldName]; exists {
		return cellType
	}
	return TypeString // Default for unknown fields
}
