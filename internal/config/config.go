package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	// All configuration options for `stui` are listed here with their defaults
	SearchDebounceInterval time.Duration = 500 * time.Millisecond
	RefreshInterval        time.Duration = 60 * time.Second
	RequestTimeout         time.Duration = 5 * time.Second
	LoadSacctDataFrom      time.Duration = 1 * time.Hour
	SlurmBinariesPath      string        = ""
	SlurmConfLocation      string        = ""
	CopyFirstColumnOnly    bool          = true
	CopiedLinesSeparator   string        = "\n"
	PartitionFilter        string        = ""
	LogLevel               int           = 2
	ShowAllColumns         bool          = false

	// Raw config options are not exposed to other modules, but pre-parsed by the config module
	rawNodeViewColumns  string = "CPULoad//CPUAlloc//CPUTot,AllocMem//RealMemory,CfgTRES,Reason,Boards"
	rawJobViewColumns   string = "UserId,JobName,RunTime,NodeList,QOS,NumCPUs,Mem"
	rawSacctViewColumns string = "QOS,Account,User,JobName,NodeList,ReqCPUS//AllocCPUS,ReqMem,Elapsed,ExitCode,ReqTRES,AllocTRES,Comment,SubmitLine"

	NodeViewColumns  *[]ColumnConfig
	JobViewColumns   *[]ColumnConfig
	SacctViewColumns *[]ColumnConfig

	// Derived config options
	SacctEnabled bool = false

	// Internal configs
	SacctMgrCurrentEntity          string = "Account" // Default starting point
	NodeStateCurrentChoice         string = ALL_CATEGORIES_OPTION
	JobStateCurrentChoice          string = ALL_CATEGORIES_OPTION
	NodeViewColumnsPartitionIndex  int
	NodeViewColumnsStateIndex      int
	JobsViewColumnsPartitionIndex  int
	JobsViewColumnsStateIndex      int
	SacctViewColumnsPartitionIndex int
	SacctViewColumnsStateIndex     int
	SacctTimeoutMultiplier         int64 = 5 // sacct can be slow, so we give it extra time

	// Cluster information
	ClusterName           string = "unknown"
	SchedulerHostName     string = "unknown"
	SchedulerSlurmVersion string = "unknown"
)

const (
	STUI_VERSION       = "0.4.3"
	KEYBOARD_SHORTCUTS = `GENERAL SHORTCUTS
1        Switch to Nodes view (scontrol)
2        Switch to Jobs view (scontrol)
3        Switch to Jobs accounting view (sacct)
4        Switch to Accounting Manager view (sacctmgr)
5        Switch to Scheduler view (sdiag)
k/j      Move selection up/down in table view
h/l      Scroll left/right in table view
Arrows   Scroll up/down/left/right in table view
?        Show this help
Ctrl+C   Exit

SHORTCUTS IN JOB/NODE VIEW
/        Open search bar to filter rows by regex, 'esc' to close, 'enter' to go back to table
p        Focus on partition selector, 'esc' to close
s        Focus on state selector, 'esc' to close
Space    Select/deselect row
y        Copy selected content (either rows, or currently open details) to clipboard
c        Open 'scontrol' prompt for selected items, or current row if no selection (opens prompt)
Enter    Show details for selected row
Esc      Close modal

ADDITIONAL SHORTCUTS IN JOBS VIEW (SCONTROL)
Ctrl+D   Open 'scancel' prompt for selected jobs, or current row if no selection

ADDITIONAL SHORTCUTS IN ACCOUNTING MANAGER VIEW (SACCTMGR)
e        Focus on Entity type selector, 'esc' to close
`

	// Below columns list fetched from Slurm 24.11.3, and are the defaults output by `scontrol` with `--details`
	ALL_OTHER_JOB_COLUMNS  = "JobName,UserId,GroupId,MCS_label,Priority,Nice,Account,QOS,WCKey,Reason,Dependency,Requeue,Restarts,BatchFlag,Reboot,ExitCode,DerivedExitCode,RunTime,TimeLimit,TimeMin,SubmitTime,EligibleTime,AccrueTime,StartTime,EndTime,Deadline,SuspendTime,SecsPreSuspend,LastSchedEval,Scheduler,AllocNode:Sid,ReqNodeList,ExcNodeList,NodeList,NumNodes,NumCPUs,NumTasks,CPUs/Task,ReqB:S:C:T,ReqTRES,AllocTRES,Socks/Node,NtasksPerN:B:S:C,CoreSpec,MinCPUsNode,MinMemoryNode,MinTmpDiskNode,Features,DelayBoot,OverSubscribe,Contiguous,Licenses,Network,Command,WorkDir,StdErr,StdIn,StdOut,TresPerTask"
	ALL_OTHER_NODE_COLUMNS = "CoresPerSocket,CPUAlloc,CPUEfctv,CPUTot,CPULoad,AvailableFeatures,ActiveFeatures,Gres,GresDrain,NodeAddr,NodeHostName,Port,RealMemory,AllocMem,FreeMem,Sockets,Boards,ThreadsPerCore,TmpDisk,Weight,Owner,MCS_label,BootTime,SlurmdStartTime,LastBusyTime,ResumeAfterTime,CfgTRES,AllocTRES,CurrentWatts,AveWatts"

	// Full list can be exported with from `sacct --helpformat | tr -s ' \n' ','`
	// The column below is a subset that excludes the fields that are always shown.
	ALL_OTHER_SACCT_COLUMNS = "AdminComment,AllocNodes,AssocID,AveCPU,AveCPUFreq,AveDiskRead,AveDiskWrite,AvePages,AveRSS,AveVMSize,BlockID,CPUTime,CPUTimeRAW,Cluster,Constraints,ConsumedEnergy,ConsumedEnergyRaw,Container,DBIndex,DerivedExitCode,ElapsedRaw,Eligible,End,Extra,FailedNode,Flags,GID,Group,JobID,Layout,Licenses,MaxDiskRead,MaxDiskReadNode,MaxDiskReadTask,MaxDiskWrite,MaxDiskWriteNode,MaxDiskWriteTask,MaxPages,MaxPagesNode,MaxPagesTask,MaxRSS,MaxRSSNode,MaxRSSTask,MaxVMSize,MaxVMSizeNode,MaxVMSizeTask,McsLabel,MinCPU,MinCPUNode,MinCPUTask,NCPUS,NNodes,NTasks,Planned,PlannedCPU,PlannedCPURAW,Priority,QOSRAW,QOSREQ,Reason,ReqCPUFreq,ReqCPUFreqGov,ReqCPUFreqMax,ReqCPUFreqMin,ReqNodes,Reservation,ReservationId,Restarts,SLUID,Start,StdErr,StdIn,StdOut,Submit,Suspended,SystemCPU,SystemComment,TRESUsageInAve,TRESUsageInMax,TRESUsageInMaxNode,TRESUsageInMaxTask,TRESUsageInMin,TRESUsageInMinNode,TRESUsageInMinTask,TRESUsageInTot,TRESUsageOutAve,TRESUsageOutMax,TRESUsageOutMaxNode,TRESUsageOutMaxTask,TRESUsageOutMin,TRESUsageOutMinNode,TRESUsageOutMinTask,TRESUsageOutTot,Timelimit,TimelimitRaw,TotalCPU,UID,UserCPU,WCKey,WCKeyID,WorkDir"

	// Certain config option names are specified as vars since they are used in other places
	CONFIG_OPTION_NAME_LOAD_SACCT_DATA_FROM = "load-sacct-data-from"

	// Log levels
	LOG_LEVEL_NONE  = 0
	LOG_LEVEL_ERROR = 1
	LOG_LEVEL_INFO  = 2
	LOG_LEVEL_DEBUG = 3

	// Misc
	ALL_CATEGORIES_OPTION = "(all)"
)

func Configure() {
	// Config flags
	flag.DurationVar(&RefreshInterval, "refresh-interval", RefreshInterval, "interval when to refetch data, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.DurationVar(&RequestTimeout, "request-timeout", RequestTimeout, "timeout setting for fetching data, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.StringVar(&SlurmBinariesPath, "slurm-binaries-path", SlurmBinariesPath, "path where Slurm binaries like 'sinfo' and 'squeue' can be found, if not in $PATH")
	flag.StringVar(&SlurmConfLocation, "slurm-conf-location", SlurmConfLocation, "path to slurm.conf for the desired cluster, if not set, fall back to SLURM_CONF env var or configless lookup if not set")
	flag.StringVar(&rawNodeViewColumns, "node-columns-config", rawNodeViewColumns, "comma-separated list of scontrol fields to show in node view, use '//' to combine columns. 'NodeName', 'Partition' and 'State' are always shown.")
	flag.StringVar(&rawJobViewColumns, "job-columns-config", rawJobViewColumns, "comma-separated list of scontrol fields to show in job view, use '//' to combine columns. 'JobId', 'Partitions' and 'JobState' are always shown.")
	flag.StringVar(&rawSacctViewColumns, "sacct-columns-config", rawSacctViewColumns, "comma-separated list of sacct fields to show in job view, use '//' to combine columns. 'JobIDRaw', 'Partitions' and 'State' are always shown.")
	flag.StringVar(&PartitionFilter, "partition", PartitionFilter, "limit views to specific partition only, leave empty to show all partitions")
	flag.BoolVar(&CopyFirstColumnOnly, "copy-first-column-only", CopyFirstColumnOnly, "if true, only copy the first column of the table to clipboard when copying")
	flag.BoolVar(&ShowAllColumns, "show-all-columns", ShowAllColumns, "if set, shows all columns for Nodes, Jobs and Accounting view Jobs, overriding other specific config")
	flag.IntVar(&LogLevel, "log-level", LogLevel, "log level, 0=none, 1=error, 2=info, 3=debug")
	flag.StringVar(&CopiedLinesSeparator, "copied-lines-separator", CopiedLinesSeparator, "string to use when separating copied lines in clipboard")
	flag.DurationVar(&LoadSacctDataFrom, CONFIG_OPTION_NAME_LOAD_SACCT_DATA_FROM, LoadSacctDataFrom, "load sacct data starting from this long ago, specify as a duration, e.g. '12h', '7d'")

	// Config flags that have been deprecated from user config
	// flag.DurationVar(&SearchDebounceInterval, "search-debounce-interval", SearchDebounceInterval, "interval to wait before searching, specify as a duration e.g. '300ms', '1s', '2m'")

	// One-shot-and-exit flags
	versionFlag := flag.Bool("version", false, "print version information and exit")
	keyboardShortcutsFlag := flag.Bool("show-keyboard-shortcuts", false, "print keyboard shortcuts and exit")

	flag.Parse()

	// Handle one shot commands
	if *versionFlag {
		fmt.Printf("stui version %s\n", STUI_VERSION)
		os.Exit(0)
	}
	if *keyboardShortcutsFlag {
		fmt.Print(KEYBOARD_SHORTCUTS)
		os.Exit(0)
	}

	// If slurm.conf location was given, ensure file exists and configure env var if appropriate
	if SlurmConfLocation != "" {
		if _, err := os.Stat(SlurmConfLocation); err != nil {
			log.Fatalf("Specified Slurm conf file cannot be found: %v", err)
		}
		err := os.Setenv("SLURM_CONF", SlurmConfLocation)
		if err != nil {
			log.Fatalf("Failed to set SLURM_CONF environment variable: %v", err)
		}
	}

	// Validate input and configs
	if RequestTimeout > RefreshInterval {
		log.Fatalf("Invalid arguments: request timeout of '%d' is longer than refresh interval of '%d'", RequestTimeout, RefreshInterval)
	}

	ComputeConfigurations()

	if err := checkIfClusterIsReachable(); err != nil {
		log.Fatalf("Failed to connect to Slurm: %v", err)
	}

	// Get scheduler info
	SchedulerHostName, ClusterName, SchedulerSlurmVersion = getSchedulerInfoWithTimeout(RequestTimeout)

	checkIfSacctMgrIsAvailable()
}

func ComputeConfigurations() {
	// Parse raw config entries
	var err error
	// Add hardcoded fields to Node
	// NodeName must be first column, as it is unique and used for selections
	// Partitions and State are used as filters and must be included.
	if ShowAllColumns {
		rawNodeViewColumns = fmt.Sprintf("NodeName,Partitions,State,%s", ALL_OTHER_NODE_COLUMNS) // Must have NodeName for selection to work!
	} else {
		rawNodeViewColumns = fmt.Sprintf("NodeName,Partitions,State,%s", rawNodeViewColumns) // Must have NodeName for selection to work!
	}
	NodeViewColumns, err = parseColumnConfigLine(rawNodeViewColumns)
	if err != nil {
		log.Fatalf("Failed to parse node column config: %v", err)
	}
	NodeViewColumnsPartitionIndex = GetColumnIndexFromColumnConfig(NodeViewColumns, "Partitions")
	NodeViewColumnsStateIndex = GetColumnIndexFromColumnConfig(NodeViewColumns, "State")

	// Add hardcoded fields to Job columns
	// JobID must be first column, as it is unique and used for selections
	// Partition and JobState are used as filters and must be included.
	// If all columns are requested, override list here.
	if ShowAllColumns {
		rawJobViewColumns = fmt.Sprintf("JobId,Partition,JobState,%s", ALL_OTHER_JOB_COLUMNS)
	} else {
		rawJobViewColumns = fmt.Sprintf("JobId,Partition,JobState,%s", rawJobViewColumns) // Must have JobID for selection to work!
	}
	JobViewColumns, err = parseColumnConfigLine(rawJobViewColumns)
	if err != nil {
		log.Fatalf("Failed to parse job column config: %v", err)
	}
	JobsViewColumnsPartitionIndex = GetColumnIndexFromColumnConfig(JobViewColumns, "Partition")
	JobsViewColumnsStateIndex = GetColumnIndexFromColumnConfig(JobViewColumns, "JobState")

	// Sacct view
	// JobID must be first column, as it is unique and used for selections
	// Partition and State are used as filters and must be included.
	// If all columns are requested, override list here.
	if ShowAllColumns {
		rawSacctViewColumns = fmt.Sprintf("JobIDRaw,Partition,State,%s", ALL_OTHER_SACCT_COLUMNS)
	} else {
		rawSacctViewColumns = fmt.Sprintf("JobIDRaw,Partition,State,%s", rawSacctViewColumns)
	}

	SacctViewColumns, err = parseColumnConfigLine(rawSacctViewColumns)
	if err != nil {
		log.Fatalf("Failed to parse sacct column config: %v", err)
	}
	SacctViewColumnsPartitionIndex = GetColumnIndexFromColumnConfig(SacctViewColumns, "Partition")
	SacctViewColumnsStateIndex = GetColumnIndexFromColumnConfig(SacctViewColumns, "State")

	// Standardise partition filter
	if PartitionFilter == "" {
		PartitionFilter = ALL_CATEGORIES_OPTION
	}
}
