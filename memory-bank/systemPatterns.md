# System Patterns: stui

## Architecture Overview

stui follows a modular architecture with clear separation between:

1. **Model Layer** (internal/model/):
   - Data providers for Slurm entities (jobs, nodes, partitions)
   - Fetchers for different Slurm commands (scontrol, sacct, etc.)
   - Parsers for command output

2. **View Layer** (internal/view/):
   - TUI components using tview library
   - Specialized selectors for different entity types
   - Modal dialogs for commands and sorting

3. **Configuration** (internal/config/):
   - Column configurations for different views
   - Settings for data fetching behavior

## Key Design Patterns

1. **Provider Pattern**:
   - Each Slurm entity type has its own provider
   - Providers implement common interfaces for data fetching
   - Example: provider_jobs.go, provider_nodes.go

2. **Command Pattern**:
   - Command modal centralizes execution of Slurm commands
   - Enables batch operations on selected entities

3. **Observer Pattern**:
   - Views subscribe to data updates from providers
   - Automatic refresh when underlying data changes

4. **Strategy Pattern**:
   - Different parsing strategies for various Slurm commands
   - Example: scontrol_fetchers.go handles multiple output formats

## Data Flow

1. User interacts with TUI view
2. View requests data from appropriate provider
3. Provider executes Slurm command and parses output
4. Parsed data returned to view for rendering
5. View updates based on new data

## Performance Considerations

- Minimizes scheduler load by:
  - Only fetching visible data
  - Implementing smart refresh intervals
  - Caching where appropriate
