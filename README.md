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
# download
sudo curl -L github.com/Antvirf/stui/releases/latest/download/stui_Linux_x86_64 -o /usr/local/bin/stui

# make it executable
sudo chmod +x /usr/local/bin/stui

# use
stui -help
```

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
         path where Slurm binaries like 'sinfo' and 'squeue' can be found (default "/usr/local/bin")
      -slurm-conf-location string
         path to slurm.conf for the desired cluster, sets 'SLURM_CONF' environment variable (default "/etc/slurm/slurm.conf")
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
    p        Focus on partition selector
    Space    Select/deselect row
    y        Copy selected rows to clipboard
    c        Run command on selected items (opens prompt)
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
