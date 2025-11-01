# Approach 2 Implementation: Domain Model First Architecture

## Overview

This document explains the implementation of Approach 2 for migrating from string parsing to JSON-based Slurm command parsing.

## Architecture

### Layer 1: Data Sources (`internal/datasource/`)

The `SlurmDataSource` interface abstracts where JSON comes from:

```go
type SlurmDataSource interface {
    FetchJobsJSON(ctx context.Context) ([]byte, error)
    FetchNodesJSON(ctx context.Context) ([]byte, error)
    FetchPartitionsJSON(ctx context.Context) ([]byte, error)
    // ... detail methods
}
```

**Implementations:**
1. **BinaryJSONSource** - Calls Slurm binaries with `--json` flag
   - `scontrol show job --json`
   - `scontrol show node --json`
   - `scontrol show partition --json`

2. **FileSource** - Reads JSON from files (testing/offline)

3. **RestAPISource** - Placeholder for future slurmrestd support

### Layer 2: JSON Parsers (`internal/parser/`)

Parsers convert raw JSON (using generated OpenAPI models) into domain models:

- `ParseJobsJSON()` - Unmarshals `V0043OpenapiJobInfoResp` → `domain.Jobs`
- `ParseNodesJSON()` - Unmarshals `V0043OpenapiNodesResp` → `domain.Nodes`
- `ParsePartitionsJSON()` - Unmarshals `V0043OpenapiPartitionResp` → `domain.Partitions`

### Layer 3: Domain Models (`internal/domain/`)

Clean, business-focused data structures independent of:
- Display format (tables, JSON, etc.)
- Data source (binary, REST, file)
- API version

```go
type Job struct {
    JobID       string
    Name        string
    User        string
    State       string
    Partition   string
    // ... 40+ fields
}
```

### Layer 4: Adapters (`internal/model/adapters.go`)

Convert domain models to TableData for the existing UI:

- `JobsToTableData(jobs, columns)` → `*TableData`
- `NodesToTableData(nodes, columns)` → `*TableData`
- `PartitionsToTableData(partitions, columns)` → `*TableData`

### Layer 5: Providers V2 (`internal/model/provider_*_v2.go`)

New providers demonstrating the pattern:

```go
func (p *JobsProviderV2) Fetch() error {
    // 1. Get JSON from data source
    jsonData, err := p.dataSource.FetchJobsJSON(ctx)
    
    // 2. Parse into domain models
    jobs, err := parser.ParseJobsJSON(jsonData)
    
    // 3. Convert to TableData for display
    tableData := JobsToTableData(jobs, config.JobViewColumns)
    
    // 4. Store both formats
    p.domainJobs = jobs        // For advanced use
    p.updateData(tableData)     // For existing UI
}
```

## Benefits

### 1. Clear Separation of Concerns
- **Data retrieval** → DataSource interface
- **Parsing** → Parser functions
- **Business logic** → Domain models
- **Presentation** → Adapters + UI

### 2. Multiple Data Sources
Easy to add new sources:
```go
// Current: Binary calls
source := datasource.NewBinaryJSONSource(timeout)

// Future: REST API
source := datasource.NewRestAPISource(url, token)

// Testing: Files
source := datasource.NewFileSource("jobs.json", "nodes.json", "partitions.json")
```

### 3. API Version Independence
- Generated OpenAPI models handle JSON structure
- Domain models shield business logic from API changes
- Adapters handle field mapping variations

### 4. Testability
Each layer can be tested independently:
- Mock data sources for parsers
- Mock parsers for providers
- Fixed domain models for adapters

### 5. Future-Proof
- slurmrestd support: Just implement `RestAPISource`
- New Slurm versions: Update OpenAPI models, adjust parsers
- New features: Extend domain models, update adapters

## Usage Example

```go
// Create data source
dataSource := datasource.NewBinaryJSONSource(config.RequestTimeout)

// Create provider
jobsProvider := model.NewJobsProviderV2(dataSource)

// Use as before
tableData := jobsProvider.FilteredData()

// OR access domain models directly
domainJobs := jobsProvider.GetDomainJobs()
for _, job := range domainJobs {
    fmt.Printf("Job %s: %s\n", job.JobID, job.State)
}
```

## Migration Path

### Phase 1: ✅ Complete
- Domain models defined
- Data source interface + implementations
- JSON parsers implemented
- Adapters created
- V2 providers as examples

### Phase 2: Future
- Add unit tests for all layers
- Create integration tests with mock data
- Gradually migrate V1 providers to use new architecture
- Maintain backward compatibility during transition

### Phase 3: Future
- Implement RestAPISource for slurmrestd
- Add configuration option to choose data source
- Performance optimization
- Full migration complete

## Files Created

```
internal/
├── domain/
│   ├── job.go              # Job domain model
│   ├── node.go             # Node domain model
│   └── partition.go        # Partition domain model
├── datasource/
│   ├── interface.go        # SlurmDataSource interface
│   ├── binary.go           # Binary JSON source
│   ├── file.go             # File source
│   └── rest.go             # REST API source (placeholder)
├── parser/
│   ├── jobs.go             # Job JSON parser
│   ├── nodes.go            # Node JSON parser
│   └── partitions.go       # Partition JSON parser
└── model/
    ├── adapters.go         # Domain → TableData converters
    ├── provider_jobs_v2.go # Jobs provider example
    ├── provider_nodes_v2.go   # Nodes provider example
    └── provider_partitions_v2.go  # Partitions provider example
```

## Key Design Decisions

1. **Why separate parsers from data sources?**
   - Data sources focus on retrieval (network, file I/O)
   - Parsers focus on transformation (JSON → domain)
   - Enables testing each independently

2. **Why domain models AND TableData?**
   - Domain models: Business logic, API-independent
   - TableData: Display format, existing UI compatibility
   - Adapters bridge the gap

3. **Why V2 providers instead of modifying V1?**
   - Demonstrates pattern without breaking existing code
   - Allows gradual migration
   - V1 and V2 can coexist during transition

4. **Why use generated OpenAPI models?**
   - Type safety when unmarshaling JSON
   - Auto-update when Slurm API changes
   - Comprehensive field coverage

## Next Steps

1. Add comprehensive unit tests
2. Create integration tests with sample JSON files
3. Add documentation for developers
4. Consider performance benchmarks (string vs JSON parsing)
5. Plan V1 → V2 migration strategy
