# `stui` - Slurm Terminal User Interface for managing clusters

*Like [k9s](https://k9scli.io/), but for Slurm clusters.* `stui` makes interacting with Slurm clusters intuitive and fast for everyone, without getting in the way of more experience users.

- List and view nodes and jobs, filter by partition and state
- Quickly search across nodes/jobs list with regular expressions
- Select multiple nodes/jobs and run `scontrol` commands on them, or copy rows to clipboard
- View individual node details (`scontrol show node` equivalent)
- View individual job details (`scontrol show job` equivalent)
- Show `sdiag` output for scheduler diagnostics
- (if Slurm accounting is enabled) Explore `sacctmgr` tables, search across rows with regular expressions
- Configure table views with specific columns/content of your choice
- Optimized to minimzie load on the Slurm scheduler by only fetching the data user is looking at. Default configs make ~1 request per minute.

`stui` requires no configuration - if you can talk to your Slurm cluster with `squeue`/`scontrol`, you can run `stui`. Several configuration options are available and detailed below.

![demo gif](./assets/demo.gif)

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
    ```txt
    Usage of ./stui:
      -copied-lines-separator string
          string to use when separating copied lines in clipboard (default "\n")
      -copy-first-column-only
          if true, only copy the first column of the table to clipboard when copying (default true)
      -default-column-width int
          minimum default width of columns in table views, if not overridden in column config (default 2)
      -job-columns-config string
          comma-separated list of scontrol fields to show in job view, suffix field name with '::<width>' to set column width, use '//' to combine columns. 'JobId', 'Partitions' and 'JobState' are always shown. (default "UserId,JobName::25,RunTime,NodeList,QOS,NumCPUs,Mem")
      -node-columns-config string
          comma-separated list of scontrol fields to show in node view, suffix field name with '::<width>' to set column width, use '//' to combine columns. 'NodeName', 'Partition' and 'State' are always shown. (default "CPULoad//CPUAlloc//CPUTot,AllocMem//RealMemory,CfgTRES::20,Reason::25,Boards")
      -partition string
          limit views to specific partition only, leave empty to show all partitions
      -quiet
          if set, do not print any log lines to console
      -refresh-interval duration
          interval when to refetch data, specify as a duration e.g. '300ms', '1s', '2m' (default 1m0s)
      -request-timeout duration
          timeout setting for fetching data, specify as a duration e.g. '300ms', '1s', '2m' (default 5s)
      -search-debounce-interval duration
          interval to wait before searching, specify as a duration e.g. '300ms', '1s', '2m' (default 50ms)
      -show-all-columns
          if set, shows all columns for both Nodes and Jobs, overriding other specific config
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
    ```txt
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
    s        Focus on state selector, 'esc' to close
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

make setup                   # install pre-commit and download Go deps
GIT_TAG=0.0.8 make release   # create release commit for given tag
```

## To-do / roadmap

- Refactor: Move search into a shared component that is shared/present across all table views, rather than something set at view level
- Refactor: Providers should either be fully aware of `config`, or not at all (currently some configs are passed in as args, and others are referred to directly)
- Feat: Basic support for plugins, similar to k9s - bash commands that can take in e.g. `$JOB_ID` or `$NODE_ID` provided by `stui`
- Feat: View stdout / tail output target of running jobs
- Feat: Add view for `sacct`: first version can use default time interval, but should be more configurable
- Feat: Ability to use `slurmrestd` / REST API instead of Slurm binaries
- Feat: Config option for which view to start app in
- Fix: highlight of currently selected row, if the cursor is on it, resets on data refresh
- Feat: support selection of objects without a clear ID column, such as certain `sacctmgtr` data like `Event`
- Feat: `TextView` or something similar to `sdiag`, so we can support `sacctmgr` text entities: `Configuration`, `Stats`

## Alternatives and why this project exists

`stui`...

- is self-contained and distributed as a single binary
- requires nothing special from Slurm to operate
- is designed for keyboard first
- is light and fast
- does not over-engineer / re-implement every feature of the Slurm binaries, but acts as a light UI to make those binaries more accessible and easier to use.

Which is in contrast to other existing projects:

- [sview](https://slurm.schedmd.com/sview.html) is the official Slurm GUI, but requires build-time configurations of Slurm, has dependencies on the OS itself
- [CLIP-HPC/SlurmCommander](https://github.com/CLIP-HPC/SlurmCommander) is an extensive project but no longer maintained, being incompatible with Slurm 23.02 and onward. Additionally, it relies on Slurm build-time configs (JSON parser) which makes it inoperable in certain environments.
- [mil-ad/stui](https://github.com/mil-ad/stui) is an older project, implementing a small terminal user interface that only supports viewing jobs, and being written in Python, needs extra work for distribution.
