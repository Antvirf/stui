package model

import (
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
)

type JobsProvider struct {
	BaseProvider[*TableData]
}

func NewJobsProvider() *JobsProvider {
	p := JobsProvider{
		BaseProvider: NewBaseProvider[*TableData](),
	}
	p.Fetch()
	return &p
}

func (p *JobsProvider) RunPeriodicRefresh(
	interval time.Duration,
	timeout time.Duration,
	callback func(),
) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		err := p.Fetch()
		if err != nil {
			callback()
		}
	}
}

func (p *JobsProvider) Fetch() error {
	// TODO: Why does this deadlock?
	// p.mu.Lock()
	// defer p.mu.Unlock()

	rawData, err := getScontrolDataWithTimeout(
		"show job --detail --all --oneliner",
		config.JobViewColumns,
		"", // TODO Partition set to blank for no filter - fix later, this shouldn't be needed here.
		"JobId=",
		config.RequestTimeout,
	)
	if err != nil {
		p.updateError(err)
		return err
	}

	p.lastUpdated = time.Now()
	p.fetchCount++

	p.updateData(rawData)
	p.length = p.data.Length()
	return nil
}

func (p *JobsProvider) FilteredData(filter string) *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()

	// Find the index of the "Partitions" field in the headers
	// TODO: Figure this out at startup, as this doesn't change
	partitionsIndex := -1
	for i, header := range *(data.Headers) {
		if header.Name == "Partition" {
			partitionsIndex = i
			break
		}
	}

	if partitionsIndex == -1 {
		return &data // Return data as-is, if partitions field isn't available
	}

	stateIndex := -1
	for i, header := range *(data.Headers) {
		if header.Name == "JobState" {
			stateIndex = i
			break
		}
	}

	var rows [][]string
	for _, row := range data.Rows {
		// Ignore row if regex filter doesn't match
		if filter != "" {
			if !strings.Contains(row[partitionsIndex], filter) {
				continue
			}
		}

		// Ignore row if State filter is set, and node doesn't match
		if stateIndex != -1 && config.JobStateCurrentChoice != "(all)" {
			if !strings.Contains(row[stateIndex], config.JobStateCurrentChoice) {
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
