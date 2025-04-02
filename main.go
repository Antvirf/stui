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
	var refreshInterval = flag.Int("refresh-interval", 30, "interval in seconds when to refetch data")
	var requestTimeout = flag.Int("request-timeout", 15, "timeout setting for fetching data")
	var debubMultiplier = flag.Int("debug-multiplier", 1, "multiplier for nodes, helpful when debugging and developing")
	// TODO: Slurm.conf location
	// TODO: Slurm binaries paths
	// TODO: Slurm RESTD option, once available
	flag.Parse()

	// Validate input and configs
	if *requestTimeout > *refreshInterval {
		log.Fatalf("Invalid arguments: request timeout of '%d' is longer than refresh interval of '%d'", *requestTimeout, *refreshInterval)
	}

	app := &view.App{
		App:             tview.NewApplication(),
		Pages:           tview.NewPages(),
		RequestTimeout:  time.Duration(*requestTimeout) * time.Second, // Must be less than refreshInterval
		DebugMultiplier: *debubMultiplier,
	}

	app.SetupViews()
	app.SetupKeybinds()
	app.StartRefresh(time.Duration(*refreshInterval) * time.Second)

	if err := app.App.SetRoot(app.MainGrid, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
