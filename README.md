# `stui` - Slurm Terminal User Interface for managing clusters

![go report](https://goreportcard.com/badge/github.com/antvirf/stui)
![loc](https://img.shields.io/badge/lines%20of%20code-4220-blue)
![size](https://img.shields.io/badge/binary%20size-5%2E4M-blue)

*Like [k9s](https://k9scli.io/), but for Slurm clusters.* `stui` makes interacting with Slurm clusters intuitive and fast for everyone, without getting in the way of more experienced users.

- List and view nodes and jobs, filter by partition and state
- Quickly search nodes/jobs lists with regular expressions across columns, sort by any column
- Select multiple nodes/jobs and run `scontrol` commands on them, run `scancel` on jobs, or copy rows to clipboard
- View individual node details (`scontrol show node` equivalent)
- View individual job details (`scontrol show job` equivalent)
- Show `sdiag` output for scheduler diagnostics
- (if Slurm accounting is enabled) Explore historical job accounting from `sacct` tables, search across rows with regular expressions, filtering by partition and state. View individual job details (`sacct -j` equivalent, with all available columns)
- (if Slurm accounting is enabled) Explore `sacctmgr` tables, search across rows with regular expressions
- Configure table views with specific columns/content of your choice
- Optimized to minimize load on the Slurm scheduler by only fetching the data user is looking at. Default configs make ~1 request per minute after initial startup.

`stui` requires no configuration - if you can talk to your Slurm cluster with `squeue`/`scontrol`, you can run `stui`. Several configuration options are available and detailed below.

![demo gif](./assets/demo.gif)

## Installation

### Install latest release for `x86_64` Linux

```bash
curl -sL stui.dev/install | sh 
```

*Installation script can be found [here](./docs/public/install) for reference.*

### Install latest release for `x86_64` Linux (manual method)

```bash
curl -sL github.com/Antvirf/stui/releases/latest/download/stui_Linux_x86_64.tar.gz | tar xzv stui
```

You can then move the binary to a more convenient location, e.g. `sudo mv stui /usr/local/bin`.

### Build + install with Go

With [`go 1.22`](https://go.dev/doc/install) or newer installed;

```bash
go install github.com/antvirf/stui@latest
alias stui=~/go/bin/stui

# root required only if you want to move the binary to a system path
sudo mv ~/go/bin/stui /usr/local/bin
```

## Usage

1. Ensure your Slurm binaries are working and you can talk to your cluster, e.g. `sdiag` shows a valid output.

2. Run `stui`. Use the `-help` flag to view arguments for additional configuration.

    <!-- REPLACE_START -->
    ```txt
    Usage of ./stui:
      -config-dir string
          path to a directory with config files (default "/home/$USER/.config/stui.d/")
      -copied-lines-separator string
          string to use when separating copied lines in clipboard (default "\n")
      -copy-first-column-only
          if true, only copy the first column of the table to clipboard when copying (default true)
      -job-columns-config string
          comma-separated list of scontrol fields to show in job view, use '//' to combine column or '++' to extend columns to full width. 'JobId', 'Partitions' and 'JobState' are always shown. (default "UserId,JobName++,RunTime,NodeList,QOS,NumCPUs,Mem")
      -load-sacct-data-from duration
          load sacct data starting from this long ago, specify as a duration, e.g. '1h', '2h'. This can be very slow on busy clusters, so use with caution. Set to 0 to not load any data from sacct. (default 30m0s)
      -log-level int
          log level, 0=none, 1=error, 2=info, 3=debug (default 2)
      -node-columns-config string
          comma-separated list of scontrol fields to show in node view, use '//' to combine column or '++' to extend columns to full width. 'NodeName', 'Partition' and 'State' are always shown. (default "CPULoad//CPUAlloc//CPUTot,AllocMem//RealMemory,CfgTRES++,Reason")
      -partition string
          limit views to specific partition only, leave empty to show all partitions
      -refresh-interval duration
          interval when to refetch data, specify as a duration e.g. '300ms', '1s', '2m' (default 1m0s)
      -request-timeout duration
          timeout setting for fetching data, specify as a duration e.g. '300ms', '1s', '2m' (default 5s)
      -sacct-columns-config string
          comma-separated list of sacct fields to show in job view, use '//' to combine columns or '++' to extend columns to full width. 'JobIDRaw', 'Partitions' and 'State' are always shown. (default "QOS,Account,User,JobName++,NodeList,ReqCPUS//AllocCPUS,ReqMem,Elapsed,ExitCode,ReqTRES,AllocTRES++,Comment++,SubmitLine++")
      -show-all-columns
          if set, shows all columns for Nodes, Jobs and Accounting view Jobs, overriding other specific config
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

3. Keyboard shortcuts within `stui`:

    <!-- REPLACE_SHORTCUTS_START -->
    ```txt
    GENERAL SHORTCUTS
    1        Switch to Nodes view (scontrol)
    2        Switch to Jobs view (scontrol)
    3        Switch to Jobs accounting view (sacct)
    4        Switch to Accounting Manager view (sacctmgr)
    5        Switch to Scheduler view (sdiag)
    k/j      Move selection up/down in table view
    h/l      Scroll left/right in table view
    Arrows   Scroll up/down/left/right in table view
    ?        Show this help
    Ctrl+R   Refresh currently visible data
    Ctrl+C   Exit
    o        Sort table by column
    
    SHORTCUTS IN JOB/NODE VIEW
    /        Open search bar to filter rows by regex, 'esc' to close, 'enter' to go back to table
    p        Focus on partition selector, 'esc' to close
    s        Focus on state selector, 'esc' to close
    Space    Select/deselect row
    y        Copy selected content (either rows, or currently open details) to clipboard
    c        Open 'scontrol' prompt for selected items, or current row if no selection (opens prompt)
    Enter    Show details for selected row
    Esc      Close modal
    
    ADDITIONAL SHORTCUTS IN JOBS VIEW (SCONTROL)
    Ctrl+D   Open 'scancel' prompt for selected jobs, or current row if no selection
    
    ADDITIONAL SHORTCUTS IN ACCOUNTING MANAGER VIEW (SACCTMGR)
    e        Focus on Entity type selector, 'esc' to close
    ```
    <!-- REPLACE_SHORTCUTS_END -->

4. Configure custom plugins/shortcuts - configure `-config-dir` argument or create a `.yaml`/`.yml` file in the default location `/home/$USER/.config/stui.d./`. Files are processed in alphabetical order. Please note that plugin configs are **concatenated**, not merged.

    - Full list of available keybinds can be found [here](https://github.com/gdamore/tcell/blob/781586687ddb57c9d44727dc9320340c4d049b11/key.go#L83-L202).
    - If several keybinds match, first plugin defined for that page takes priority.
    - Plugins are processed after existing keybinds, and cannot override the defaults.
    - Any column in a given table view is available for use, following standard [Go template](https://pkg.go.dev/text/template) syntax.

    <!-- REPLACE_CONFIG_EXAMPLE_START -->
    ```yaml
    plugins:
      - name: Sstat a job
        # Available pages: `nodes`, `jobs`, `sacct`, `sacctmgr`
        activePage: jobs
        shortcut: "Ctrl-S"
        # Any column of a particular view can be used in a command template
        command: sstat {{.JobId}}
        # Whether to execute command immediately rather than open a prompt. Default is false.
        executeImmediately: true
        # Closes prompt immediately once command is executed. Default is false.
        # Only applies if `executeImmediately` is true.
        closePromptAfterExecute: false
    
      - name: Open logs from a remote HTTP server
        activePage: jobs
        shortcut: "Ctrl-U"
        command: firefox "https://localhost:8080/{{.JobId}}" > /dev/null &
        executeImmediately: true
        closePromptAfterExecute: true
    
      - name: Check node disk usage
        activePage: nodes
        shortcut: "Ctrl-S"
        command: ssh {{.NodeName}} 'df -h /'
    ```
    <!-- REPLACE_CONFIG_EXAMPLE_END -->

## Developing `stui`

The below helpers configure a locally running cluster with `888` virtual nodes across several partitions to help work on `stui` with realistic data. This builds Slurm from scratch, so refer to [Slurm docs on build dependencies.](https://slurm.schedmd.com/quickstart_admin.html#manual_build)

```bash
make build-cluster      # build Slurm with required options
make config-cluster     # copy mock config to /etc/slurm/
make run-cluster        # start `slurmctld` and `slurmd`
make setup-sacct        # set up sacct
make launch-jobs        # launch few hundred sleep jobs
make stop-cluster       # stop cluster

make setup                      # install pre-commit and download Go deps
GIT_TAG=0.0.8 make gh-release   # create release commit for given tag
```

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

## FAQ / Troubleshooting

### Strange colors on tmux

This is likely the result of `tmux` defaulting to a different colour mode than the terminal emulator being used to run it is expecting. You can usually fix this by adding `export TERM=screen-256color` to your shell RC files.
