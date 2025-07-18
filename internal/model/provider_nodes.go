package model

import (
	"github.com/antvirf/stui/internal/config"
)

type NodesProvider struct {
	BaseProvider[*TableData]
}

func NewNodesProvider() *NodesProvider {
	p := NodesProvider{
		BaseProvider: BaseProvider[*TableData]{},
	}
	p.Fetch()
	return &p
}

func (p *NodesProvider) Fetch() error {
	// Compute column widths on first fetch only
	computeColumnWidths := false
	if p.lastUpdated.IsZero() {
		computeColumnWidths = true
	}
	rawData, err := getScontrolDataWithTimeout(
		"show node --detail --all --oneliner",
		config.NodeViewColumns,
		config.RequestTimeout,
		computeColumnWidths,
	)
	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
	return nil
}

func (p *NodesProvider) FilteredData() *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.ApplyFilters(
		map[int]string{
			config.NodeViewColumnsStateIndex:     config.NodeStateCurrentChoice,
			config.NodeViewColumnsPartitionIndex: config.PartitionFilter,
		},
	)
}
