# Product Context: stui

## Purpose

stui is a terminal user interface for Slurm cluster management, designed to:

- Provide intuitive and fast interaction with Slurm clusters
- Serve both novice and experienced users effectively
- Minimize scheduler load while providing real-time data

## Key User Problems Solved

1. **Complexity of Slurm commands**: Abstracts away complex `scontrol`, `sacct`, and `sacctmgr` commands
2. **Information overload**: Provides focused views of cluster data with filtering capabilities
3. **Slow workflows**: Enables batch operations on multiple nodes/jobs with single commands
4. **Monitoring challenges**: Offers real-time views of nodes, jobs, and scheduler diagnostics

## Core User Workflows

1. Cluster monitoring:
   - View node states and job statuses
   - Check scheduler diagnostics via sdiag
2. Job management:
   - Search/filter jobs
   - Cancel/modify jobs in bulk
3. Node management:
   - Search/filter nodes
   - Modify node states in bulk
4. Historical analysis:
   - Explore accounting data (when available)
   - Review sacctmgr configurations

## Key Features

- Real-time node/job views with filtering
- Bulk operations via command modal
- Accounting data exploration
- Minimal configuration requirements
- Optimized scheduler load
