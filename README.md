# S9S

*Like [k9s](https://k9scli.io/), but for Slurm*

Terminal User Interface (TUI) for viewing and managing Slurm nodes and jobs.

## TODO

- Generic table model, similar to k9s: Able to show any tables
- Generic DAO/view approach similar to k9s for data updates
- Persistent footer
  - Tabs at the bottom to switch views (nodes / jobs / scheduler)
  - Current version
  - Path to conf
  - Target cluster hostname/port, conn info
  - Show 'data as of'

## Target feature list for MVP

- Node list view: Overall data, state, partitions
  - Node detail view: press enter to show all node details
- Jobs view: Show live jobs, overall state
  - Jobs detail view: Press enter to go show job detail (popup pane): perhaps from scontrol?
- Scheduler view: Just pull sdiag and show info
-

## Nicer to haves

- General "namespace" type limit for jobs/nodes
