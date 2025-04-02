# `stui` - Slurm terminal user interface

*Like [k9s](https://k9scli.io/), but for Slurm*

Terminal User Interface (TUI) for viewing and managing Slurm nodes and jobs.

```bash
go install github.com/antvirf/stui@latest
sudo mv ~/go/bin/stui /usr/bin
```

## To-do

- Proper code structure
  - Configurability with `flag`
  - Data fetchers into their own internal package
- General ability to 'select' rows (both jobs / nodes), first feature just `yank` the data
- Control commands: Set node state and reason for all selected nodes
- Control commands: Cancel jobs / Send to top of queue for all selected jobs
- Selector/limit by partition across both job and node views
- Ability to use `slurmrestd` / REST API instead of Slurm binaries
