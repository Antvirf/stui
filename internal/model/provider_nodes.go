package model

import (
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
)

type NodesProvider struct {
	BaseProvider[*TableData]
}

func NewNodesProvider() *NodesProvider {
	p := NodesProvider{
		BaseProvider: NewBaseProvider[*TableData](),
	}
	p.Fetch()
	return &p
}

func (p *NodesProvider) RunPeriodicRefresh(
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

func (p *NodesProvider) Fetch() error {
	// TODO: Why does this deadlock?
	// p.mu.Lock()
	// defer p.mu.Unlock()

	rawData, err := getScontrolDataWithTimeout(
		"show node --detail --all --oneliner",
		config.NodeViewColumns,
		"", // TODO Partition set to blank for no filter - fix later, this shouldn't be needed here.
		"NodeName=",
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

func (p *NodesProvider) FilteredData(filter string) *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()

	// Find the index of the "Partitions" field in the headers
	// TODO: Figure this out at startup, as this doesn't change
	partitionsIndex := -1
	for i, header := range *(data.Headers) {
		if header.Name == "Partitions" {
			partitionsIndex = i
			break
		}
	}

	stateIndex := -1
	for i, header := range *(data.Headers) {
		if header.Name == "State" {
			stateIndex = i
			break
		}
	}

	if partitionsIndex == -1 {
		return &data // Return data as-is, if partitions field isn't available
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
		if stateIndex != -1 && config.NodeStateCurrentChoice != "(all)" {
			if !strings.Contains(row[stateIndex], config.NodeStateCurrentChoice) {
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
