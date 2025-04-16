package model

import (
	"time"

	"github.com/antvirf/stui/internal/config"
)

type PartitionsProvider struct {
	BaseProvider[*TableData]
}

func NewPartitionsProvider() *PartitionsProvider {
	p := PartitionsProvider{
		BaseProvider: NewBaseProvider[*TableData](),
	}
	p.Fetch()
	return &p
}

func (p *PartitionsProvider) RunPeriodicRefresh(
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

func (p *PartitionsProvider) Fetch() error {
	// TODO: Why does this deadlock?
	// p.mu.Lock()
	// defer p.mu.Unlock()

	rawData, err := getScontrolDataWithTimeout(
		"show partitions --detail --all --oneliner",
		&[]config.ColumnConfig{{Name: "PartitionName"}},
		"", // No filter
		"PartitionName=",
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

// PartitionsProvider data does not have a categorical filter, so this just returns the current data.
func (p *PartitionsProvider) FilteredData(filter string) *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()
	return &data
}
