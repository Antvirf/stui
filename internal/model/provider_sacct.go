package model

import (
	"time"

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
	// Compute column widths on first fetch only
	computeColumnWidths := false
	if p.lastUpdated.IsZero() {
		computeColumnWidths = true
	}
	rawData, err := getSacctDataSinceWithTimeout(
		config.LoadSacctDataFrom,
		config.SacctViewColumns,
		time.Duration(
			config.SacctTimeoutMultiplier*config.RequestTimeout.Milliseconds(),
		)*time.Millisecond,
		computeColumnWidths,
	)

	// Empty table data is returned in case of error, so this is always valid to do
	p.updateData(rawData)

	if err != nil {
		p.updateError(err)
		return err
	}
	return nil
}

func (p *SacctProvider) FilteredData() *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.ApplyFilters(
		map[int]string{
			config.SacctViewColumnsStateIndex:     config.JobStateCurrentChoice,
			config.SacctViewColumnsPartitionIndex: config.PartitionFilter,
		},
	)
}
