package model

import (
	"context"
	
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/datasource"
	"github.com/antvirf/stui/internal/domain"
	"github.com/antvirf/stui/internal/parser"
)

// PartitionsProviderV2 is the new provider using domain models and data sources
type PartitionsProviderV2 struct {
	BaseProvider[*TableData]
	dataSource        datasource.SlurmDataSource
	domainPartitions domain.Partitions
}

// NewPartitionsProviderV2 creates a new partitions provider using the new architecture
func NewPartitionsProviderV2(dataSource datasource.SlurmDataSource) *PartitionsProviderV2 {
	p := &PartitionsProviderV2{
		BaseProvider: BaseProvider[*TableData]{
			data: EmptyTableData(),
		},
		dataSource: dataSource,
	}
	p.Fetch()
	return p
}

// Fetch retrieves partitions using the data source and parser
func (p *PartitionsProviderV2) Fetch() error {
	FetchCounter.increment()
	
	// Compute column widths on first fetch only
	computeColumnWidths := p.lastUpdated.IsZero()

	// Step 1: Get JSON from data source
	ctx := context.Background()
	jsonData, err := p.dataSource.FetchPartitionsJSON(ctx)
	if err != nil {
		p.updateError(err)
		return err
	}

	// Step 2: Parse JSON into domain models
	partitions, err := parser.ParsePartitionsJSON(jsonData)
	if err != nil {
		p.updateError(err)
		return err
	}

	// Store domain models for potential future use
	p.domainPartitions = partitions

	// Step 3: Convert domain models to TableData for display
	// Use a simple column config for partitions
	columns := &[]config.ColumnConfig{{RawName: "PartitionName", DisplayName: "PartitionName"}}
	tableData := PartitionsToTableData(partitions, columns)
	
	// Compute column widths if needed
	if computeColumnWidths {
		p.computeColumnWidths(tableData)
	}

	p.updateData(tableData)
	return nil
}

// FilteredData returns the data (partitions typically don't have filters)
func (p *PartitionsProviderV2) FilteredData() *TableData {
	return p.Data()
}

// GetDomainPartitions returns the domain models directly (for advanced use cases)
func (p *PartitionsProviderV2) GetDomainPartitions() domain.Partitions {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.domainPartitions
}

// computeColumnWidths updates column widths based on content
func (p *PartitionsProviderV2) computeColumnWidths(tableData *TableData) {
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
