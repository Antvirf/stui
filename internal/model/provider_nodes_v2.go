package model

import (
	"context"
	
	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/datasource"
	"github.com/antvirf/stui/internal/domain"
	"github.com/antvirf/stui/internal/parser"
)

// NodesProviderV2 is the new provider using domain models and data sources
type NodesProviderV2 struct {
	BaseProvider[*TableData]
	dataSource  datasource.SlurmDataSource
	domainNodes domain.Nodes
}

// NewNodesProviderV2 creates a new nodes provider using the new architecture
func NewNodesProviderV2(dataSource datasource.SlurmDataSource) *NodesProviderV2 {
	p := &NodesProviderV2{
		BaseProvider: BaseProvider[*TableData]{
			data: EmptyTableData(),
		},
		dataSource: dataSource,
	}
	p.Fetch()
	return p
}

// Fetch retrieves nodes using the data source and parser
func (p *NodesProviderV2) Fetch() error {
	FetchCounter.increment()
	
	// Compute column widths on first fetch only
	computeColumnWidths := p.lastUpdated.IsZero()

	// Step 1: Get JSON from data source
	ctx := context.Background()
	jsonData, err := p.dataSource.FetchNodesJSON(ctx)
	if err != nil {
		p.updateError(err)
		return err
	}

	// Step 2: Parse JSON into domain models
	nodes, err := parser.ParseNodesJSON(jsonData)
	if err != nil {
		p.updateError(err)
		return err
	}

	// Store domain models for potential future use
	p.domainNodes = nodes

	// Step 3: Convert domain models to TableData for display
	tableData := NodesToTableData(nodes, config.NodeViewColumns)
	
	// Compute column widths if needed
	if computeColumnWidths {
		p.computeColumnWidths(tableData)
	}

	p.updateData(tableData)
	return nil
}

// FilteredData applies filters to the table data
func (p *NodesProviderV2) FilteredData() *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.ApplyFilters(
		map[int]string{
			config.NodeViewColumnsStateIndex:     config.NodeStateCurrentChoice,
			config.NodeViewColumnsPartitionIndex: config.PartitionFilter,
		},
	)
}

// GetDomainNodes returns the domain models directly (for advanced use cases)
func (p *NodesProviderV2) GetDomainNodes() domain.Nodes {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.domainNodes
}

// computeColumnWidths updates column widths based on content
func (p *NodesProviderV2) computeColumnWidths(tableData *TableData) {
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
