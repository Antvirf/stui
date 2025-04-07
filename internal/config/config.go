package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.design/x/clipboard"
)

var (
	// All configuration options for `stui` are listed here with their defaults
	SearchDebounceInterval time.Duration = 50 * time.Millisecond
	RefreshInterval        time.Duration = 15 * time.Second
	RequestTimeout         time.Duration = 4 * time.Second
	SlurmBinariesPath      string        = ""
	SlurmConfLocation      string        = ""
	NodeViewColumns        string        = "NodeName,Partitions,State,CPUTot,RealMemory,CPULoad,Reason,Sockets,CoresPerSocket,ThreadsPerCore,Gres"
	JobViewColumns         string        = "JobId,UserId,Partition,JobName,JobState,RunTime,NodeList"
	CopyFirstColumnOnly    bool          = true
	CopiedLinesSeparator   string        = "\n"
	PartitionFilter        string        = ""

	// Derived config options
	NodeStatusField    string = "State"
	JobStatusField     string = "JobState"
	ClipboardAvailable bool   = false
)

const (
	STUI_VERSION       = "0.0.2"
	KEYBOARD_SHORTCUTS = `General Shortcuts
1        Switch to Nodes view
2        Switch to Jobs view
3        Switch to Scheduler view
Up/Down  Move selection up/down
k/j      Move selection up/down
?        Show this help
Ctrl+C   Exit

Shortcuts in Job/Node panes
/        Open search bar to filter rows by regex
p        Focus on partition selector, 'esc' to close
Space    Select/deselect row
y        Copy selected content (either rows, or currently open details) to clipboard
c        Run command on selected items, or on current row if no selection (opens prompt)
Enter    Show details for selected row
Esc      Close modal
`
)

func Configure() {
	// Config flags
	flag.DurationVar(&SearchDebounceInterval, "search-debounce-interval", SearchDebounceInterval, "interval to wait before searching, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.DurationVar(&RefreshInterval, "refresh-interval", RefreshInterval, "interval when to refetch data, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.DurationVar(&RequestTimeout, "request-timeout", RequestTimeout, "timeout setting for fetching data, specify as a duration e.g. '300ms', '1s', '2m'")
	flag.StringVar(&SlurmBinariesPath, "slurm-binaries-path", SlurmBinariesPath, "path where Slurm binaries like 'sinfo' and 'squeue' can be found, if not in $PATH")
	flag.StringVar(&SlurmConfLocation, "slurm-conf-location", SlurmConfLocation, "path to slurm.conf for the desired cluster, if not set, fall back to SLURM_CONF env var or configless lookup if not set")
	flag.StringVar(&NodeViewColumns, "node-view-columns", NodeViewColumns, "comma-separated list of scontrol fields to show in node view")
	flag.StringVar(&JobViewColumns, "job-view-columns", JobViewColumns, "comma-separated list of scontrol fields to show in job view")
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
		fmt.Println(KEYBOARD_SHORTCUTS)
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

	// Compute derived configs
	if !strings.Contains(NodeViewColumns, NodeStatusField) {
		NodeStatusField = ""
	}
	if !strings.Contains(JobViewColumns, JobStatusField) {
		JobStatusField = ""
	}

	// We set clipboard available only if this works
	// See package docs at https://github.com/golang-design/clipboard for more detail
	err := clipboard.Init()
	if err == nil {
		ClipboardAvailable = true
	}
}
