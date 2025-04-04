package config

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	// All configuration options for `stui` are listed here
	SearchDebounceInterval time.Duration
	RefreshInterval        time.Duration
	RequestTimeout         time.Duration
	DebugMultiplier        int
	SlurmBinariesPath      string
	SlurmConfLocation      string
	SlurmRestdAddress      string
)

func Configure() {
	flag.DurationVar(&SearchDebounceInterval, "search-debounce-interval", 50, "interval in milliseconds to wait before searching")
	flag.DurationVar(&RefreshInterval, "refresh-interval", 15, "interval in seconds when to refetch data")
	flag.DurationVar(&RequestTimeout, "request-timeout", 4, "timeout setting for fetching data")
	flag.IntVar(&DebugMultiplier, "debug-multiplier", 1, "multiplier for nodes and jobs, helpful when debugging and developing")
	flag.StringVar(&SlurmBinariesPath, "slurm-binaries-path", "/usr/local/bin", "path where Slurm binaries like 'sinfo' and 'squeue' can be found")
	flag.StringVar(&SlurmConfLocation, "slurm-conf-location", "/etc/slurm/slurm.conf", "path to slurm.conf for the desired cluster, sets 'SLURM_CONF' environment variable")
	flag.StringVar(&SlurmRestdAddress, "slurm-restd-address", "", "URI for Slurm REST API if available, including protocol and port")
	flag.Parse()

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
	log.Printf("WARNING: flag value is currently unimplemented: slurm-binaries-path='%s'", SlurmBinariesPath)
	log.Printf("WARNING: flag value is currently unimplemented: slurm-restd-address='%s'", SlurmRestdAddress)
}
