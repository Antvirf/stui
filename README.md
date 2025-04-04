# `stui` - Slurm terminal user interface

*Like [k9s](https://k9scli.io/), but for Slurm*

Terminal User Interface (TUI) for viewing and managing Slurm nodes and jobs.

## Features

- List and view nodes and jobs, quickly filter list with regexp
- View individual node details (`scontrol show node` equivalent)
- View individual job details (`scontrol show job` equivalent)
- Show `sdiag` output

## Installation

With [`go 1.22`](https://go.dev/doc/install) or newer installed;

```bash
go install github.com/antvirf/stui@latest
sudo mv ~/go/bin/stui /usr/bin
```

## Usage

1. Ensure your Slurm binaries are working and you can talk to your cluster, e.g. `sdiag` shows a valid output.

2. Run `stui` / `go run main.go` in the repo. See `-help` for arguments.

    <!-- REPLACE_START -->
    ```
    Usage of ./stui:
      -copied-lines-separator string
        	string to use when separating copied lines in clipboard (default "\n")
      -copy-first-column-only
        	if true, only copy the first column of the table to clipboard when copying (default true)
      -job-view-columns string
        	comma-separated list of scontrol fields to show in job view (default "JobId,UserId,Partition,JobName,JobState,RunTime,NodeList")
      -node-view-columns string
        	comma-separated list of scontrol fields to show in node view (default "NodeName,Partitions,State,CPUTot,RealMemory,CPULoad,Reason,Sockets,CoresPerSocket,ThreadsPerCore,Gres")
      -partition string
        	limit views to specific partition only, leave empty to show all partitions
      -refresh-interval duration
        	interval in seconds when to refetch data (default 15ns)
      -request-timeout duration
        	timeout setting for fetching data (default 4ns)
      -search-debounce-interval duration
        	interval in milliseconds to wait before searching (default 50ns)
      -show-keyboard-shortcuts
        	print keyboard shortcuts and exit
      -slurm-binaries-path string
        	path where Slurm binaries like 'sinfo' and 'squeue' can be found (default "/usr/local/bin")
      -slurm-conf-location string
        	path to slurm.conf for the desired cluster, sets 'SLURM_CONF' environment variable (default "/etc/slurm/slurm.conf")
      -slurm-restd-address string
        	URI for Slurm REST API if available, including protocol and port
      -version
        	print version information and exit
    ```
    <!-- REPLACE_END -->

3. Keyboard shorcuts within `stui`

    <!-- REPLACE_SHORTCUTS_START -->
    ```
    General Shortcuts
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
    
    ```
    <!-- REPLACE_SHORTCUTS_END -->

## Developing `stui`

The below helpers configure a locally running cluster with `888` virtual nodes across several partitions to help work on `stui` with realistic data.

```bash
make build-cluster      # build Slurm with required options
make config-cluster     # copy mock config to /etc/slurm/
make run-cluster        # start `slurmctld` and `slurmd`
make launch-jobs        # launch few hundred sleep jobs
make stop-cluster       # stop cluster

make setup              # install pre-commit and download Go deps
```

## To-do

- Control commands: Set node state and reason for all selected nodes
- Control commands: Cancel jobs / Send to top of queue for all selected jobs
- Improve handling of sdiag/other calls if no scheduler available - by default they hang for a long time
- Ability to use `slurmrestd` / REST API instead of Slurm binaries
