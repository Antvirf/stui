package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	// All configuration options for `stui` are listed here
	SearchDebounceInterval time.Duration
	RefreshInterval        time.Duration
	RequestTimeout         time.Duration
	SlurmBinariesPath      string
	SlurmConfLocation      string
	SlurmRestdAddress      string
	NodeViewColumns        string
	JobViewColumns         string
	CopyFirstColumnOnly    bool
	CopiedLinesSeparator   string
	PartitionFilter        string
)

const (
	KEYBOARD_SHORTCUTS = `General Shortcuts
1        Switch to Nodes view
2        Switch to Jobs view
3        Switch to Scheduler view
Up/Down  Move selection up/down
k/j      Move selection up/down
?        Show this help

Shortcuts in Job/Node panes
/        Open search bar to filter rows by regex
p        Focus on partition selector (up/down to navigate, 'Enter' to select)
Space    Select/deselect row
y        Copy selected rows to clipboard
Enter    Show details for selected row
Esc      Close modal
`
)

func Configure() {
	// Config flags
	flag.DurationVar(&SearchDebounceInterval, "search-debounce-interval", 50, "interval in milliseconds to wait before searching")
	flag.DurationVar(&RefreshInterval, "refresh-interval", 15, "interval in seconds when to refetch data")
	flag.DurationVar(&RequestTimeout, "request-timeout", 4, "timeout setting for fetching data")
	flag.StringVar(&SlurmBinariesPath, "slurm-binaries-path", "/usr/local/bin", "path where Slurm binaries like 'sinfo' and 'squeue' can be found")
	flag.StringVar(&SlurmConfLocation, "slurm-conf-location", "/etc/slurm/slurm.conf", "path to slurm.conf for the desired cluster, sets 'SLURM_CONF' environment variable")
	flag.StringVar(&SlurmRestdAddress, "slurm-restd-address", "", "URI for Slurm REST API if available, including protocol and port")
	flag.StringVar(&NodeViewColumns, "node-view-columns", "NodeName,Partitions,State,CPUTot,RealMemory,CPULoad,Reason,Sockets,CoresPerSocket,ThreadsPerCore,Gres", "comma-separated list of scontrol fields to show in node view")
	flag.StringVar(&JobViewColumns, "job-view-columns", "JobId,UserId,Partition,JobName,JobState,RunTime,NodeList", "comma-separated list of scontrol fields to show in job view")
	flag.StringVar(&PartitionFilter, "partition", "", "limit views to specific partition only, leave empty to show all partitions")
	flag.BoolVar(&CopyFirstColumnOnly, "copy-first-column-only", true, "if true, only copy the first column of the table to clipboard when copying")
	flag.StringVar(&CopiedLinesSeparator, "copied-lines-separator", "\n", "string to use when separating copied lines in clipboard")

	// One-shot-and-exit flags
	versionFlag := flag.Bool("version", false, "print version information and exit")
	keyboardShortcutsFlag := flag.Bool("show-keyboard-shortcuts", false, "print keyboard shortcuts and exit")

	flag.Parse()

	// Handle oneshots
	if *versionFlag {
		fmt.Println("stui version 0.0.1")
		os.Exit(0)
	}
	if *keyboardShortcutsFlag {
		fmt.Println(KEYBOARD_SHORTCUTS)
		os.Exit(0)
	}

	// Set up durations with correct units
	SearchDebounceInterval = SearchDebounceInterval * time.Millisecond
	RefreshInterval = RefreshInterval * time.Second
	RequestTimeout = RequestTimeout * time.Second

	// If slurm.conf location was given, ensure file exists and configure env var if appropriate
	if SlurmConfLocation != "" {
		if _, err := os.Stat(SlurmConfLocation); err != nil {
			log.Fatalf("Specified Slurm conf file cannot be found: %v", err)
		}

		if err := os.Setenv("SLURM_CONF", SlurmConfLocation); err != nil {
			log.Fatalf("Failed to set SLURM_CONF environment variable: %v", err)
		}
	}

	// Validate input and configs
	if RequestTimeout > RefreshInterval {
		log.Fatalf("Invalid arguments: request timeout of '%d' is longer than refresh interval of '%d'", RequestTimeout, RefreshInterval)
	}

	// Warnings about incomplete features
	log.Printf("WARNING: flag value is currently unimplemented: slurm-restd-address='%s'", SlurmRestdAddress)
}
