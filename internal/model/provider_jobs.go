package model

import (
	"github.com/antvirf/stui/internal/config"
)

type JobsProvider struct {
	BaseProvider[*TableData]
}

func NewJobsProvider() *JobsProvider {
	p := JobsProvider{
		BaseProvider: BaseProvider[*TableData]{},
	}
	p.Fetch()
	return &p
}

func (p *JobsProvider) Fetch() error {
	rawData, err := getScontrolDataWithTimeout(
		"show job --detail --all --oneliner",
		config.JobViewColumns,
		config.RequestTimeout,
	)
	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
	return nil
}

func (p *JobsProvider) FilteredData() *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.data.ApplyFilters(
		map[int]string{
			config.JobsViewColumnsStateIndex:     config.JobStateCurrentChoice,
			config.JobsViewColumnsPartitionIndex: config.PartitionFilter,
		},
	)
}
