# `stui` - Slurm Terminal User Interface for managing clusters

*Like [k9s](https://k9scli.io/), but for Slurm clusters*

- List and view nodes and jobs, across all partitions or a specific partition
- Quickly filter list nodes/jobs list with regular expressions
- Select multiple nodes/jobs and run `scontrol` commands on them
- View individual node details (`scontrol show node` equivalent)
- View individual job details (`scontrol show job` equivalent)
- Show `sdiag` output

![](./assets/demo.gif)

## Installation

### Install latest release for `x86_64` Linux

```bash
curl -sL github.com/Antvirf/stui/releases/latest/download/stui_Linux_x86_64.tar.gz | tar xzv stui
```

You can then move the binary to a more convenient location, e.g. `sudo mv stui /usr/bin/local`.

### Build + install with Go (does not require root)

With [`go 1.22`](https://go.dev/doc/install) or newer installed;

```bash
go install github.com/antvirf/stui@latest
alias stui=~/go/bin/stui

# root required only if you want to move the binary to a system path
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
        	interval when to refetch data, specify as a duration e.g. '300ms', '1s', '2m' (default 15s)
      -request-timeout duration
        	timeout setting for fetching data, specify as a duration e.g. '300ms', '1s', '2m' (default 4s)
      -search-debounce-interval duration
        	interval to wait before searching, specify as a duration e.g. '300ms', '1s', '2m' (default 50ms)
      -show-keyboard-shortcuts
        	print keyboard shortcuts and exit
      -slurm-binaries-path string
        	path where Slurm binaries like 'sinfo' and 'squeue' can be found, if not in $PATH
      -slurm-conf-location string
        	path to slurm.conf for the desired cluster, if not set, fall back to SLURM_CONF env var or configless lookup if not set
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
    Ctrl+C   Exit
    
    Shortcuts in Job/Node panes
    /        Open search bar to filter rows by regex
    p        Focus on partition selector, 'esc' to close
    Space    Select/deselect row
    y        Copy selected content (either rows, or currently open details) to clipboard
    c        Run command on selected items, or on current row if no selection (opens prompt)
    Enter    Show details for selected row
    Esc      Close modal
    
    ```
    <!-- REPLACE_SHORTCUTS_END -->

## Developing `stui`

The below helpers configure a locally running cluster with `888` virtual nodes across several partitions to help work on `stui` with realistic data. This builds Slurm from scratch, so refer to [Slurm docs on build dependencies.](https://slurm.schedmd.com/quickstart_admin.html#manual_build)

```bash
make build-cluster      # build Slurm with required options
make config-cluster     # copy mock config to /etc/slurm/
make run-cluster        # start `slurmctld` and `slurmd`
make launch-jobs        # launch few hundred sleep jobs
make stop-cluster       # stop cluster

make setup              # install pre-commit and download Go deps
```

## To-do

- Feat: Footer should contain overall node/job counts by state
- Feat: View stdout / tail output target of running jobs
- Improve handling of sdiag/other calls if no scheduler available - by default they hang for a long time, perhaps check at launch that a cluster is reachable
- Ability to use `slurmrestd` / REST API instead of Slurm binaries
