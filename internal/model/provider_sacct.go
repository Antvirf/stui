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
	rawData, err := getSacctDataSinceWithTimeout(
		config.LoadSacctDataFrom,
		config.SacctViewColumns,
		time.Duration(
			config.SacctTimeoutMultiplier*config.RequestTimeout.Milliseconds(),
		)*time.Millisecond,
	)
	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
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
