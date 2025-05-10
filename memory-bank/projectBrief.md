# `stui` project brief

## Overarching goal

*Like [k9s](https://k9scli.io/), but for Slurm clusters.* `stui` makes interacting with Slurm clusters intuitive and fast for everyone, without getting in the way of more experienced users.

## Existing features

- List and view nodes and jobs, filter by partition and state
- Quickly search across nodes/jobs list with regular expressions
- Select multiple nodes/jobs and run `scontrol` commands on them, run `scancel` on jobs, or copy rows to clipboard
- View individual node details (`scontrol show node` equivalent)
- View individual job details (`scontrol show job` equivalent)
- Show `sdiag` output for scheduler diagnostics
- (if Slurm accounting is enabled) Explore historical job accounting from `sacct` tables, search across rows with regular expressions, filtering by partition and state. This view is not refreshed automatically.
- (if Slurm accounting is enabled) Explore `sacctmgr` tables, search across rows with regular expressions
- Configure table views with specific columns/content of your choice
- Optimized to minimize load on the Slurm scheduler by only fetching the data user is looking at. Default configs make ~1 request per minute after initial startup.

`stui` requires no configuration - if you can talk to your Slurm cluster with `squeue`/`scontrol`, you can run `stui`. Several configuration options are available and detailed below.

## To-do / roadmap

- Feat: Sorting: Ctrl+s, open a pane to select one of the visible columns
- Feat: `sacct` view further features: ability to search for jobs beyond currently loaded data, and/or ability to change time range within the view itself
- Feat: `sstat` option for running jobs (returns tabular data, tbc how to do that nicely)
- Feat: Summary stats in top middle pane: Node and job states
- Feat: Basic support for plugins, similar to k9s - bash commands that can take in e.g. `$JOB_ID` or `$NODE_ID` provided by `stui`
- Feat: Ability to use `slurmrestd` / REST API instead of Slurm binaries
- Feat: Config option for which view to start app in
- Fix: highlight of currently selected row, if the cursor is on it, resets on data refresh
- Feat: support selection of objects without a clear ID column, such as certain `sacctmgr` data like `Event`
- Feat: `TextView` or something similar to `sdiag`, so we can support `sacctmgr` text entities: `Configuration`, `Stats`
- Refactor: clean up where/how commands are defined, currently has some repetition
- Refactor: keep track of 'connection state' to scheduler: right now if a connection is lost, switching between views becomes slow due to timeout + `FetchIfStaleAndRender`, which tries to query the scheduler on every refresh
