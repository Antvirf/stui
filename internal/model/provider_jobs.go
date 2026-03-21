package model

import (
	"github.com/antvirf/stui/internal/config"
)

type JobsProvider struct {
	BaseProvider[*TableData]
}

func NewJobsProvider(loadData bool) *JobsProvider {
	p := JobsProvider{
		BaseProvider: BaseProvider[*TableData]{
			data: EmptyTableData(),
		},
	}
	if loadData {
		p.Fetch()
	}
	return &p
}

func (p *JobsProvider) Fetch() error {
	// Compute column widths on first fetch only
	computeColumnWidths := false
	if p.lastUpdated.IsZero() {
		computeColumnWidths = true
	}
	rawData, err := getScontrolDataWithTimeout(
		"show job --detail --all",
		config.JobViewColumns,
		config.RequestTimeout,
		computeColumnWidths,
		parseScontrolJobsOutput,
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
