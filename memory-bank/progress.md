# Project Progress: stui

## Completed Features

- Basic node/job listing and viewing
- Filtering by partition and state
- Regex search across nodes/jobs
- Bulk scontrol commands
- Individual node/job details
- sdiag view for scheduler diagnostics
- Accounting views (sacct, sacctmgr)
- Table view configuration
- Optimized scheduler load

## Roadmap Items

- sacct view enhancements:
  - Time range adjustment
  - Extended search capabilities
- sstat integration for running jobs
- Summary stats shown in the top middle bar for each table: e.g. overall nodes / drained/ down /alloc /idle split etc.
- Plugin system for custom commands
- slurmrestd/REST API support
- Startup view configuration

## Known Issues

- Selected row highlighting resets on refresh
- Connection state not properly tracked
- Some sacctmgr entities lack ID columns, and row selection does not work for them
- Text views needed for certain non-tabular sacctmgr data
- Command definition duplication
