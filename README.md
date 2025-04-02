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

    ```
    $ stui -help
    Usage of ~/go/bin/stui:
      -debug-multiplier int
            multiplier for nodes, helpful when debugging and developing (default 1)
      -refresh-interval int
            interval in seconds when to refetch data (default 30)
      -request-timeout int
            timeout setting for fetching data (default 15)
      -slurm-binaries-path string
            path where Slurm binaries like 'sinfo' and 'squeue' can be found (default "/usr/local/bin")
      -slurm-conf-location string
            path to slurm.conf for the desired cluster (default "/etc/slurm/slurm.conf")
      -slurm-restd-address string
            URI for Slurm REST API if available, including protocol and port

    ```

## To-do

- Proper code structure
  - Data fetchers into their own internal package, use config options from main
  - Separate out view and data models cleanly
- General ability to 'select' rows (both jobs / nodes), first feature just `yank` the data
- Control commands: Set node state and reason for all selected nodes
- Control commands: Cancel jobs / Send to top of queue for all selected jobs
- Selector/limit by partition across both job and node views
- Ability to use `slurmrestd` / REST API instead of Slurm binaries
