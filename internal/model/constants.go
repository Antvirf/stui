package model

var (
	// https://slurm.schedmd.com/sinfo.html
	SCONTROL_NODE_STATES = []string{
		"(all)", // Default choice, special case
		"ALLOC",
		"ALLOCATED",
		"BLOCKED",
		"CLOUD",
		"COMP",
		"COMPLETING",
		"DOWN",
		"DRAIN",
		"DRAINED",
		"DRAINING",
		"FAIL",
		"FUTURE",
		"FUTR",
		"IDLE",
		"MAINT",
		"MIX",
		"MIXED",
		"NO_RESPOND",
		"NPC",
		"PERFCTRS",
		"PLANNED",
		"POWER_DOWN",
		"POWERING_DOWN",
		"POWERED_DOWN",
		"POWERING_UP",
		"REBOOT_ISSUED",
		"REBOOT_REQUESTED",
		"RESV",
		"RESERVED",
		"UNK",
		"UNKNOWN",
	}

	// https://slurm.schedmd.com/job_state_codes.html#states
	SCONTROL_JOB_STATES = []string{
		"(all)", // Default choice, special case
		"BOOT_FAIL",
		"CANCELLED",
		"COMPLETED",
		"DEADLINE",
		"FAILED",
		"NODE_FAIL",
		"OUT_OF_MEMORY",
		"PENDING",
		"PREEMPTED",
		"RUNNING",
		"SUSPENDED",
		"TIMEOUT",
	}

	SACCTMGR_TABLE_ENTITIES = []string{
		"Account",
		"Association",
		"Cluster",
		"Event",
		"Federation",
		// "Problem", // Requires admin
		"QOS",
		"Resource",
		"Reservation",
		// "RunAwayJobs", // Requires operator perms
		"Transaction",
		"TRES",
		"User",
		// "WCKey", // Requires admin
	}
	SACCTMGR_TEXT_ENTITIES = []string{
		"Configuration",
		// "Stats", // Requires admin
	}

	SACCTMGR_ENTITY_COLUMN_CONFIGS = map[string]string{
		"Account":     "Org,Account,Descr",
		"Association": "Cluster,Account,User,Partition,Share,QOS,Def QOS,Priority,GrpJobs,GrpTRES,GrpSubmit,GrpWall,GrpTRESMins,MaxJobs,MaxTRES,MaxTRESPerNode,MaxSubmit,MaxWall,MaxTRESMins,GrpTRESRunMins",
		"Cluster":     "Cluster,ControlHost,ControlPort,RPC,Share,QOS,Def QOS,GrpJobs,GrpTRES,GrpSubmit,MaxJobs,MaxTRES,MaxSubmit,MaxWall",
		"Event":       "Cluster,NodeName,TimeStart,TimeEnd,State,Reason,User",
		"Federation":  "Cluster,NodeName,TimeStart,TimeEnd,State,Reason,User",
		"QOS":         "Name,Priority,GraceTime,Preempt,PreemptExemptTime,PreemptMode,Flags,UsageThres,UsageFactor,GrpTRES,GrpTRESMins,GrpTRESRunMins,GrpJobs,GrpSubmit,GrpWall,MaxTRES,MaxTRESPerNode,MaxTRESMins,MaxWall,MaxTRESPU,MaxJobsPU,MaxSubmitPU,MaxTRESPA,MaxTRESRunMinsPA,MaxTRESRunMinsPU,MaxJobsPA,MaxSubmitPA,MinTRES",
		"Resource":    "Name,Server,Type,Count,LastConsumed,Allocated,ServerType,Flags",
		"Reservation": "Cluster,Name,TRES,TimeStart,TimeEnd,UnusedWall",
		"Transaction": "Time,Action,Actor,Where,Info",
		"TRES":        "Type,Name,ID",
		"User":        "User, Def Acct, Def WCKey, Admin",
	}
)
