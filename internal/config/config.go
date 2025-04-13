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
	SearchDebounceInterval time.Duration = 50 * time.Millisecond
	RefreshInterval        time.Duration = 15 * time.Second
	RequestTimeout         time.Duration = 4 * time.Second
	SlurmBinariesPath      string        = ""
	SlurmConfLocation      string        = ""
	CopyFirstColumnOnly    bool          = true
	CopiedLinesSeparator   string        = "\n"
	PartitionFilter        string        = ""
	DefaultColumnWidth     int           = 2

	// Raw config options are not exposed to other modules, but pre-parsed by the config module
	rawNodeViewColumns string = "NodeName,Partitions:15,State,CPUAlloc//CPUTot,AllocMem//RealMemory,CfgTRES:20,Reason:25,Boards"
	rawJobViewColumns  string = "JobId,Partition,UserId,JobName:25,JobState,RunTime,NodeList,QOS,NumCPUs,Mem"
	NodeViewColumns    *[]ColumnConfig
	JobViewColumns     *[]ColumnConfig

	// Derived config options
	NodeStatusField string = "State"
	JobStatusField  string = "JobState"
	SacctEnabled    bool   = false
)

const (
	STUI_VERSION       = "0.0.6"
	KEYBOARD_SHORTCUTS = `General Shortcuts
1        Switch to Nodes view
2        Switch to Jobs view
3        Switch to Scheduler view
4        Switch to Accounting Manager view (if sacctmgr is available)
k/j      Move selection up/down in table view
h/l      Scroll left/right in table view
Arrows   Scroll up/down/left/right in table view
?        Show this help
Ctrl+C   Exit

Shortcuts in Job/Node view
/        Open search bar to filter rows by regex, 'esc' to close, 'enter' to go back to table
p        Focus on partition selector, 'esc' to close
Space    Select/deselect row
y        Copy selected content (either rows, or currently open details) to clipboard
c        Run command on selected items, or on current row if no selection (opens prompt)
Enter    Show details for selected row
Esc      Close modal

Additional shortcuts in Accounting Manager view
e        Focus on Entity type selector, 'esc' to close
`
)

func Configure() {
	// Config flags
	flag.DurationVar(&SearchDebounceInterval, "search-debounce-interval", SearchDebounceInterval, "interval to wait before searching, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.DurationVar(&RefreshInterval, "refresh-interval", RefreshInterval, "interval when to refetch data, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.DurationVar(&RequestTimeout, "request-timeout", RequestTimeout, "timeout setting for fetching data, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.StringVar(&SlurmBinariesPath, "slurm-binaries-path", SlurmBinariesPath, "path where Slurm binaries like 'sinfo' and 'squeue' can be found, if not in $PATH")
	flag.StringVar(&SlurmConfLocation, "slurm-conf-location", SlurmConfLocation, "path to slurm.conf for the desired cluster, if not set, fall back to SLURM_CONF env var or configless lookup if not set")
	flag.StringVar(&rawNodeViewColumns, "node-columns-config", rawNodeViewColumns, "comma-separated list of scontrol fields to show in node view, suffix field name with ':<width>' to set column width, use '//' to combine columns.")
	flag.StringVar(&rawJobViewColumns, "job-columns-config", rawJobViewColumns, "comma-separated list of scontrol fields to show in job view, suffix field name with ':<width>' to set column width, use '//' to combine columns.")
	flag.IntVar(&DefaultColumnWidth, "default-column-width", DefaultColumnWidth, "minimum default width of columns in table views, if not overridden in column config")
	flag.StringVar(&PartitionFilter, "partition", PartitionFilter, "limit views to specific partition only, leave empty to show all partitions")
	flag.BoolVar(&CopyFirstColumnOnly, "copy-first-column-only", CopyFirstColumnOnly, "if true, only copy the first column of the table to clipboard when copying")
	flag.StringVar(&CopiedLinesSeparator, "copied-lines-separator", CopiedLinesSeparator, "string to use when separating copied lines in clipboard")

	// One-shot-and-exit flags
	versionFlag := flag.Bool("version", false, "print version information and exit")
	keyboardShortcutsFlag := flag.Bool("show-keyboard-shortcuts", false, "print keyboard shortcuts and exit")

	flag.Parse()

	// Handle oneshots
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
	checkIfSacctMgrIsAvailable()
}

// Compute configs, assuming inputs are all provided and valid
func ComputeConfigurations() {
	// Compute derived configs
	if !strings.Contains(rawNodeViewColumns, NodeStatusField) {
		NodeStatusField = ""
	}
	if !strings.Contains(rawJobViewColumns, JobStatusField) {
		JobStatusField = ""
	}

	// Parse raw config entries
	var err error
	NodeViewColumns, err = parseColumnConfigLine(rawNodeViewColumns)
	if err != nil {
		log.Fatalf("Failed to parse node column config: %v", err)
	}
	JobViewColumns, err = parseColumnConfigLine(rawJobViewColumns)
	if err != nil {
		log.Fatalf("Failed to parse job column config: %v", err)
	}
}
