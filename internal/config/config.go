package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var (
	// All configuration options for `stui` are listed here with their defaults
	SearchDebounceInterval time.Duration = 500 * time.Millisecond
	RefreshInterval        time.Duration = 60 * time.Second
	RequestTimeout         time.Duration = 5 * time.Second
	LoadSacctDataFrom      time.Duration = 12 * time.Hour
	SlurmBinariesPath      string        = ""
	SlurmConfLocation      string        = ""
	CopyFirstColumnOnly    bool          = true
	CopiedLinesSeparator   string        = "\n"
	PartitionFilter        string        = ""
	DefaultColumnWidth     int           = 2
	LogLevel               int           = 2
	ShowAllColumns         bool          = false

	// Raw config options are not exposed to other modules, but pre-parsed by the config module
	rawNodeViewColumns  string = "CPULoad//CPUAlloc//CPUTot,AllocMem//RealMemory,CfgTRES::20,Reason::25,Boards"
	rawJobViewColumns   string = "UserId,JobName::25,RunTime,NodeList,QOS,NumCPUs,Mem"
	rawSacctViewColumns string = "JobName,AllocCPUS,ReqMem,Elapsed,ExitCode,ReqTRES,AllocTRES,ReqCPUFreqMin,ReqCPUFreqMax,ReqCPUFreqGov"
	NodeViewColumns     *[]ColumnConfig
	JobViewColumns      *[]ColumnConfig
	SacctViewColumns    *[]ColumnConfig

	// Derived config options
	NodeStatusField string = "State"
	JobStatusField  string = "JobState"
	SacctEnabled    bool   = false

	// Internal configs
	SacctMgrCurrentEntity          string = "Account" // Default starting point
	NodeStateCurrentChoice         string = "(all)"
	JobStateCurrentChoice          string = "(all)"
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
	STUI_VERSION       = "0.3.1"
	KEYBOARD_SHORTCUTS = `General Shortcuts
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

Shortcuts in Job/Node view
/        Open search bar to filter rows by regex, 'esc' to close, 'enter' to go back to table
p        Focus on partition selector, 'esc' to close
s        Focus on state selector, 'esc' to close
Space    Select/deselect row
y        Copy selected content (either rows, or currently open details) to clipboard
c        Open 'scontrol' prompt for selected items, or current row if no selection (opens prompt)
Enter    Show details for selected row
Esc      Close modal

Additional shortcuts in Jobs view
Ctrl+D   Open 'scancel' prompt for selected jobs, or current row if no selection

Additional shortcuts in Accounting Manager view
e        Focus on Entity type selector, 'esc' to close
`

	// Below columns list fetched from Slurm 24.11.3
	ALL_OTHER_JOB_COLUMNS   = "JobName,UserId,GroupId,MCS_label,Priority,Nice,Account,QOS,WCKey,Reason,Dependency,Requeue,Restarts,BatchFlag,Reboot,ExitCode,DerivedExitCode,RunTime,TimeLimit,TimeMin,SubmitTime,EligibleTime,AccrueTime,StartTime,EndTime,Deadline,SuspendTime,SecsPreSuspend,LastSchedEval,Scheduler,AllocNode:Sid,ReqNodeList,ExcNodeList,NodeList,NumNodes,NumCPUs,NumTasks,CPUs/Task,ReqB:S:C:T,ReqTRES,AllocTRES,Socks/Node,NtasksPerN:B:S:C,CoreSpec,MinCPUsNode,MinMemoryNode,MinTmpDiskNode,Features,DelayBoot,OverSubscribe,Contiguous,Licenses,Network,Command,WorkDir,StdErr,StdIn,StdOut,TresPerTask"
	ALL_OTHER_NODE_COLUMNS  = "CoresPerSocket,CPUAlloc,CPUEfctv,CPUTot,CPULoad,AvailableFeatures,ActiveFeatures,Gres,GresDrain,NodeAddr,NodeHostName,Port,RealMemory,AllocMem,FreeMem,Sockets,Boards,ThreadsPerCore,TmpDisk,Weight,Owner,MCS_label,BootTime,SlurmdStartTime,LastBusyTime,ResumeAfterTime,CfgTRES,AllocTRES,CurrentWatts,AveWatts"
	ALL_OTHER_SACCT_COLUMNS = "MaxVMSize,MaxVMSizeNode,MaxVMSizeTask,AveVMSize,MaxRSS,MaxRSSNode,MaxRSSTask,AveRSS,MaxPages,MaxPagesNode,MaxPagesTask,AvePages,MinCPU,MinCPUNode,MinCPUTask,AveCPU,NTasks,AveCPUFreq,ConsumedEnergy,MaxDiskRead,MaxDiskReadNode,MaxDiskReadTask,AveDiskRead,MaxDiskWrite,MaxDiskWriteNode,MaxDiskWriteTask,AveDiskWrite,TRESUsageInAve,TRESUsageInMax,TRESUsageInMaxNode,TRESUsageInMaxTask,TRESUsageInMin,TRESUsageInMinNode,TRESUsageInMinTask,TRESUsageInTot,TRESUsageOutMax,TRESUsageOutMaxNode,TRESUsageOutMaxTask,TRESUsageOutAve,TRESUsageOutTot"

	// Certain config option names are specified as vars since they are used in other places
	CONFIG_OPTION_NAME_LOAD_SACCT_DATA_FROM = "load-sacct-data-from"

	// Log levels
	LOG_LEVEL_NONE  = 0
	LOG_LEVEL_ERROR = 1
	LOG_LEVEL_INFO  = 2
	LOG_LEVEL_DEBUG = 3
)

func Configure() {
	// Config flags
	flag.DurationVar(&SearchDebounceInterval, "search-debounce-interval", SearchDebounceInterval, "interval to wait before searching, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.DurationVar(&RefreshInterval, "refresh-interval", RefreshInterval, "interval when to refetch data, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.DurationVar(&RequestTimeout, "request-timeout", RequestTimeout, "timeout setting for fetching data, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.StringVar(&SlurmBinariesPath, "slurm-binaries-path", SlurmBinariesPath, "path where Slurm binaries like 'sinfo' and 'squeue' can be found, if not in $PATH")
	flag.StringVar(&SlurmConfLocation, "slurm-conf-location", SlurmConfLocation, "path to slurm.conf for the desired cluster, if not set, fall back to SLURM_CONF env var or configless lookup if not set")
	flag.StringVar(&rawNodeViewColumns, "node-columns-config", rawNodeViewColumns, "comma-separated list of scontrol fields to show in node view, suffix field name with '::<width>' to set column width, use '//' to combine columns. 'NodeName', 'Partition' and 'State' are always shown.")
	flag.StringVar(&rawJobViewColumns, "job-columns-config", rawJobViewColumns, "comma-separated list of scontrol fields to show in job view, suffix field name with '::<width>' to set column width, use '//' to combine columns. 'JobId', 'Partitions' and 'JobState' are always shown.")
	flag.StringVar(&rawSacctViewColumns, "sacct-columns-config", rawSacctViewColumns, "comma-separated list of sacct fields to show in job view, suffix field name with '::<width>' to set column width, use '//' to combine columns. 'JobIDRaw', 'Partitions' and 'State' are always shown.")
	flag.IntVar(&DefaultColumnWidth, "default-column-width", DefaultColumnWidth, "minimum default width of columns in table views, if not overridden in column config")
	flag.StringVar(&PartitionFilter, "partition", PartitionFilter, "limit views to specific partition only, leave empty to show all partitions")
	flag.BoolVar(&CopyFirstColumnOnly, "copy-first-column-only", CopyFirstColumnOnly, "if true, only copy the first column of the table to clipboard when copying")
	flag.BoolVar(&ShowAllColumns, "show-all-columns", ShowAllColumns, "if set, shows all columns for both Nodes and Jobs, overriding other specific config")
	flag.IntVar(&LogLevel, "log-level", LogLevel, "log level, 0=none, 1=error, 2=info, 3=debug")
	flag.StringVar(&CopiedLinesSeparator, "copied-lines-separator", CopiedLinesSeparator, "string to use when separating copied lines in clipboard")
	flag.DurationVar(&LoadSacctDataFrom, CONFIG_OPTION_NAME_LOAD_SACCT_DATA_FROM, LoadSacctDataFrom, "load sacct data starting from this long ago, specify as a duration, e.g. '12h', '7d'")

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
	// Compute derived configs
	if !strings.Contains(rawNodeViewColumns, NodeStatusField) {
		NodeStatusField = ""
	}
	if !strings.Contains(rawJobViewColumns, JobStatusField) {
		JobStatusField = ""
	}
	if !strings.Contains(rawSacctViewColumns, JobStatusField) {
		JobStatusField = ""
	}

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
	NodeViewColumnsPartitionIndex = 1
	NodeViewColumnsStateIndex = 2

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
	JobsViewColumnsPartitionIndex = 1
	JobsViewColumnsStateIndex = 2

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
	SacctViewColumnsPartitionIndex = 1
	SacctViewColumnsStateIndex = 2

}
