package model

import (
	"strings"

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
		"JobId=",
		config.RequestTimeout,
	)
	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
	return nil
}

func (p *JobsProvider) FilteredData(filter string) *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()

	var rows [][]string
	for _, row := range data.Rows {
		// Ignore row if partition filter doesn't match
		if filter != "" {
			if !strings.Contains(row[config.JobsViewColumnsPartitionIndex], filter) {
				continue
			}
		}

		// Ignore row if the state filer doesn't match
		if config.JobStateCurrentChoice != "(all)" {
			if !strings.Contains(row[config.JobsViewColumnsStateIndex], config.JobStateCurrentChoice) {
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
