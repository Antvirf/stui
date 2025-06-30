package model

import (
	"github.com/antvirf/stui/internal/config"
)

type PartitionsProvider struct {
	BaseProvider[*TableData]
}

func NewPartitionsProvider() *PartitionsProvider {
	p := PartitionsProvider{
		BaseProvider: BaseProvider[*TableData]{},
	}
	p.Fetch()
	return &p
}

func (p *PartitionsProvider) Fetch() error {
	rawData, err := getScontrolDataWithTimeout(
		"show partitions --detail --all --oneliner",
		&[]config.ColumnConfig{{RawName: "PartitionName", DisplayName: "PartitionName"}},
		config.RequestTimeout,
		false, // Don't compute column widths, doesn't matter here.
	)

	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
	return nil
}

// PartitionsProvider data does not have a categorical filter, so this just returns the current data.
func (p *PartitionsProvider) FilteredData() *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()
	return &data
}
