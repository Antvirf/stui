# `stui` - Slurm Terminal User Interface for managing clusters

*Like [k9s](https://k9scli.io/), but for Slurm clusters.* `stui` makes interacting with Slurm clusters intuitive and fast for everyone, without getting in the way of more experience users.

- List and view nodes and jobs, across all partitions or a specific partition
- Quickly filter list nodes/jobs list with regular expressions
- Select multiple nodes/jobs and run `scontrol` commands on them, or copy rows to clipboard
- Configure table views with specific columns/content of your choice
- View individual node details (`scontrol show node` equivalent)
- View individual job details (`scontrol show job` equivalent)
- Show `sdiag` output for scheduler diagnostics
- (if Slurm accounting is enabled) Explore `sacctmgr` tables, filter rows with regular expressions

`stui` requires no configuration - if you can talk to your Slurm cluster with `squeue`/`scontrol`, you can run `stui`. Several configuration options are available and detailed below.

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

2. Run `stui`. Use the `-help` flag to view arguments for additional configuration.

    <!-- REPLACE_START -->
    ```
    Usage of ./stui:
      -copied-lines-separator string
        	string to use when separating copied lines in clipboard (default "\n")
      -copy-first-column-only
        	if true, only copy the first column of the table to clipboard when copying (default true)
      -default-column-width int
        	minimum default width of columns in table views, if not overridden in column config (default 2)
      -job-columns-config string
        	comma-separated list of scontrol fields to show in job view, suffix field name with ':<width>' to set column width, use '//' to combine columns. (default "JobId,Partition,UserId,JobName:25,JobState,RunTime,NodeList,QOS,NumCPUs,Mem")
      -node-columns-config string
        	comma-separated list of scontrol fields to show in node view, suffix field name with ':<width>' to set column width, use '//' to combine columns. (default "NodeName,Partitions:15,State,CPUAlloc//CPUTot,AllocMem//RealMemory,CfgTRES:20,Reason:25,Boards")
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
    ```
    <!-- REPLACE_SHORTCUTS_END -->

## Developing `stui`

The below helpers configure a locally running cluster with `888` virtual nodes across several partitions to help work on `stui` with realistic data. This builds Slurm from scratch, so refer to [Slurm docs on build dependencies.](https://slurm.schedmd.com/quickstart_admin.html#manual_build)

```bash
make build-cluster      # build Slurm with required options
make config-cluster     # copy mock config to /etc/slurm/
make run-cluster        # start `slurmctld` and `slurmd`
make setup-sacct        # set up sacct
make launch-jobs        # launch few hundred sleep jobs
make stop-cluster       # stop cluster

make setup              # install pre-commit and download Go deps
```

## To-do

- Separation of internal data fetched vs. data used to render table
- Break apart app.go into smaller pieces
  - Initialization
  - Layout
  - Data refreshes
  - Refactor search bar / overall grid layout logic. Quite gross atm
- Proper error propagation, so that individual data update calls to e.g. permission denied resources will fail gracefully and with clear error messages (e.g. for `sacctmgr` some commands may be off limits)
- Feat: View stdout / tail output target of running jobs
- Improve handling of sdiag/other calls if no scheduler available - by default they hang for a long time, perhaps check at launch that a cluster is reachable
- Add view for `sacct`: first version can use default time interval, but should be more configurable
- Ability to use `slurmrestd` / REST API instead of Slurm binaries
- Config option for which view to start app in
- Fix: highlight of currently selected row, if the cursor is on it, resets on data refresh
