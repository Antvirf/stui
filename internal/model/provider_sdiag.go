package model

import (
	"github.com/antvirf/stui/internal/config"
)

type SdiagProvider struct {
	BaseProvider[*TextData]
}

func NewSdiagProvider(loadData bool) *SdiagProvider {
	p := SdiagProvider{
		BaseProvider: BaseProvider[*TextData]{},
	}
	if loadData {
		p.Fetch()
	} else {
		p.data = &TextData{Data: ""}
	}
	return &p
}

func (p *SdiagProvider) Fetch() error {
	rawData, err := getSdiagWithTimeout(config.RequestTimeout)

	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(&TextData{Data: rawData})
	return nil
}

// SdiagProvider data does not have a categorical filter, so this just returns the current data.
func (p *SdiagProvider) FilteredData() *TextData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()
	return &data
}
