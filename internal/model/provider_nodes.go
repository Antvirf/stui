package model

import (
	"strings"

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
	rawData, err := getScontrolDataWithTimeout(
		"show node --detail --all --oneliner",
		config.NodeViewColumns,
		"NodeName=",
		config.RequestTimeout,
	)
	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
	return nil
}

func (p *NodesProvider) FilteredData(filter string) *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()

	var rows [][]string
	for _, row := range data.Rows {
		// Ignore row if partition filter doesn't match
		if filter != "" {
			if !strings.Contains(row[config.NodeViewColumnsPartitionIndex], filter) {
				continue
			}
		}

		// Ignore row if state filtr doesn't match
		if config.NodeStateCurrentChoice != "(all)" {
			if !strings.Contains(row[config.NodeViewColumnsStateIndex], config.NodeStateCurrentChoice) {
				continue
			}
		}
		rows = append(rows, row)
	}

	return &TableData{
		Headers: data.Headers,
		Rows:    rows,
	}
}
