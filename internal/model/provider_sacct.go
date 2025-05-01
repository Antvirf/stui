package model

import (
	"strings"

	"github.com/antvirf/stui/internal/config"
)

type SacctProvider struct {
	BaseProvider[*TableData]
}

func NewSacctProvider() *SacctProvider {
	p := SacctProvider{
		BaseProvider: BaseProvider[*TableData]{},
	}
	p.Fetch()
	return &p
}

func (p *SacctProvider) Fetch() error {
	rawData, err := getSacctDataSince(
		config.LoadSacctDataFrom,
	)
	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
	return nil
}

func (p *SacctProvider) FilteredData(filter string) *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()

	var rows [][]string
	for _, row := range data.Rows {
		// Ignore row if partition filter doesn't match
		if filter != "" {
			if !strings.Contains(row[config.SacctViewColumnsPartitionIndex], filter) {
				continue
			}
		}

		// Ignore row if state filter doesn't match
		if config.JobStateCurrentChoice != "(all)" {
			if !strings.Contains(row[config.SacctViewColumnsStateIndex], config.JobStateCurrentChoice) {
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
