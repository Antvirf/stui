package model

import (
	"context"
	
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/datasource"
	"github.com/antvirf/stui/internal/domain"
	"github.com/antvirf/stui/internal/parser"
)

// JobsProviderV2 is the new provider using domain models and data sources
type JobsProviderV2 struct {
	BaseProvider[*TableData]
	dataSource datasource.SlurmDataSource
	domainJobs domain.Jobs
}

// NewJobsProviderV2 creates a new jobs provider using the new architecture
func NewJobsProviderV2(dataSource datasource.SlurmDataSource) *JobsProviderV2 {
	p := &JobsProviderV2{
		BaseProvider: BaseProvider[*TableData]{
			data: EmptyTableData(),
		},
		dataSource: dataSource,
	}
	p.Fetch()
	return p
}

// Fetch retrieves jobs using the data source and parser
func (p *JobsProviderV2) Fetch() error {
	FetchCounter.increment()
	
	// Compute column widths on first fetch only
	computeColumnWidths := p.lastUpdated.IsZero()

	// Step 1: Get JSON from data source
	ctx := context.Background()
	jsonData, err := p.dataSource.FetchJobsJSON(ctx)
	if err != nil {
		p.updateError(err)
		return err
	}

	// Step 2: Parse JSON into domain models
	jobs, err := parser.ParseJobsJSON(jsonData)
	if err != nil {
		p.updateError(err)
		return err
	}

	// Store domain models for potential future use
	p.domainJobs = jobs

	// Step 3: Convert domain models to TableData for display
	tableData := JobsToTableData(jobs, config.JobViewColumns)
	
	// Compute column widths if needed
	if computeColumnWidths {
		p.computeColumnWidths(tableData)
	}

	p.updateData(tableData)
	return nil
}

// FilteredData applies filters to the table data
func (p *JobsProviderV2) FilteredData() *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.ApplyFilters(
		map[int]string{
			config.JobsViewColumnsStateIndex:     config.JobStateCurrentChoice,
			config.JobsViewColumnsPartitionIndex: config.PartitionFilter,
		},
	)
}

// GetDomainJobs returns the domain models directly (for advanced use cases)
func (p *JobsProviderV2) GetDomainJobs() domain.Jobs {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.domainJobs
}

// computeColumnWidths updates column widths based on content
func (p *JobsProviderV2) computeColumnWidths(tableData *TableData) {
	if tableData.Headers == nil {
		return
	}
	
	for i := range *tableData.Headers {
		col := &(*tableData.Headers)[i]
		maxWidth := len(col.DisplayName)
		
		for _, row := range tableData.Rows {
			if i < len(row) {
				cellWidth := len(row[i])
				if cellWidth > maxWidth {
					maxWidth = cellWidth
				}
			}
		}
		
		// Cap at maximum column width
		if maxWidth > config.MaximumColumnWidth {
			maxWidth = config.MaximumColumnWidth
		}
		
		col.Width = maxWidth
	}
}
