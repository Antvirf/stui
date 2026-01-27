package model

// FieldTypeMap maps Slurm field names to their expected types.
// Unknown fields default to TypeString.
var FieldTypeMap = map[string]CellType{
	// Integer fields - Nodes
	"CPUTot":    TypeInteger,
	"CPUAlloc":  TypeInteger,
	"NodeCount": TypeInteger,

	// Integer fields - Jobs
	"JobId":          TypeInteger,
	"NumCPUs":        TypeInteger,
	"ReqCPUS":        TypeInteger,
	"AllocCPUS":      TypeInteger,
	"NumNodes":       TypeInteger,
	"NumTasks":       TypeInteger,
	"Priority":       TypeInteger,
	"ArrayJobId":     TypeInteger,
	"ArrayTaskId":    TypeInteger,
	"Nice":           TypeInteger,
	"Requeue":        TypeInteger,
	"Restarts":       TypeInteger,
	"BatchFlag":      TypeInteger,
	"Reboot":         TypeInteger,
	"Contiguous":     TypeInteger,
	"MinCPUsNode":    TypeInteger,
	"MinTmpDiskNode": TypeInteger,
	"SecsPreSuspend": TypeInteger,

	// Float fields
	"CPULoad": TypeFloat,

	// Memory fields
	"AllocMem":      TypeMemory,
	"RealMemory":    TypeMemory,
	"ReqMem":        TypeMemory,
	"Memory":        TypeMemory,
	"Mem":           TypeMemory,
	"MinMemoryNode": TypeMemory,

	// Duration fields (time spans)
	"RunTime":     TypeDuration,
	"TimeLimit":   TypeDuration,
	"Elapsed":     TypeDuration,
	"TimeMin":     TypeDuration,
	"DelayBoot":   TypeDuration,
	"SuspendTime": TypeDuration,

	// Timestamp fields (absolute points in time)
	"SubmitTime":    TypeTimestamp,
	"StartTime":     TypeTimestamp,
	"EndTime":       TypeTimestamp,
	"EligibleTime":  TypeTimestamp,
	"AccrueTime":    TypeTimestamp,
	"Deadline":      TypeTimestamp,
	"LastSchedEval": TypeTimestamp,

	// String fields (explicit for documentation)
	// Node fields
	"NodeName":       TypeString,
	"State":          TypeString,
	"Gres":           TypeString,
	"ActiveFeatures": TypeString,

	// Job fields
	"JobState":   TypeString,
	"Partition":  TypeString,
	"Partitions": TypeString,
	"Account":    TypeString,
	"User":       TypeString,
	"UserId":     TypeString,
	"GroupId":    TypeString,
	"JobName":    TypeString,
	"NodeList":   TypeString,
	"QOS":        TypeString,
	"Comment":    TypeString,
	"Reason":     TypeString,
	"Dependency": TypeString,
	"Scheduler":  TypeString,
	"MCS_label":  TypeString,
	"WCKey":      TypeString,

	// Resource specification fields (complex formats)
	"CfgTRES":     TypeString,
	"AllocTRES":   TypeString,
	"ReqTRES":     TypeString,
	"TresPerTask": TypeString,
	"SubmitLine":  TypeString,

	// Exit codes (format: "code:signal")
	"ExitCode":        TypeString,
	"DerivedExitCode": TypeString,
}

// GetFieldType returns the type for a field name, defaulting to TypeString for unknown fields
func GetFieldType(fieldName string) CellType {
	if cellType, exists := FieldTypeMap[fieldName]; exists {
		return cellType
	}
	return TypeString // Default for unknown fields
}
