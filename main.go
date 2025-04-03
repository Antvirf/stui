package main

import (
	"flag"
	"log"
	"time"

	"github.com/antvirf/stui/internal/view"
	"github.com/rivo/tview"
)

func main() {
	// Flags
	var searchDebounceInterval = flag.Int("search-debounce-interval", 50, "interval in milliseconds to wait before searching")
	var refreshInterval = flag.Int("refresh-interval", 30, "interval in seconds when to refetch data")
	var requestTimeout = flag.Int("request-timeout", 15, "timeout setting for fetching data")
	var debubMultiplier = flag.Int("debug-multiplier", 1, "multiplier for nodes, helpful when debugging and developing")
	var slurmBinariesPath = flag.String("slurm-binaries-path", "/usr/local/bin", "path where Slurm binaries like 'sinfo' and 'squeue' can be found")
	var slurmConfLocation = flag.String("slurm-conf-location", "/etc/slurm/slurm.conf", "path to slurm.conf for the desired cluster")
	var slurmRestdAddress = flag.String("slurm-restd-address", "", "URI for Slurm REST API if available, including protocol and port")
	flag.Parse()

	// Validate input and configs
	if *requestTimeout > *refreshInterval {
		log.Fatalf("Invalid arguments: request timeout of '%d' is longer than refresh interval of '%d'", *requestTimeout, *refreshInterval)
	}
	log.Printf("WARNING: flag value is currently unimplemented: slurm-binaries-path='%s'", *slurmBinariesPath)
	log.Printf("WARNING: flag value is currently unimplemented: slurm-conf-location='%s'", *slurmConfLocation)
	log.Printf("WARNING: flag value is currently unimplemented: slurm-restd-address='%s'", *slurmRestdAddress)

	app := &view.App{
		App:                    tview.NewApplication(),
		Pages:                  tview.NewPages(),
		RequestTimeout:         time.Duration(*requestTimeout) * time.Second, // Must be less than refreshInterval
		DebugMultiplier:        *debubMultiplier,
		SearchDebounceInterval: time.Duration(*searchDebounceInterval) * time.Millisecond,
	}

	app.SetupViews()
	app.SetupKeybinds()
	app.StartRefresh(time.Duration(*refreshInterval) * time.Second)

	if err := app.App.SetRoot(app.MainGrid, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
