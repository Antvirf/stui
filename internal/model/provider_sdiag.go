package model

import (
	"github.com/antvirf/stui/internal/config"
	"time"
)

type SdiagProvider struct {
	BaseProvider[*TextData]
}

func NewSdiagProvider(loadData bool) *SdiagProvider {
	p := SdiagProvider{
		BaseProvider: BaseProvider[*TextData]{},
	}
	p.data = &TextData{Data: ""}
	if loadData {
		p.Fetch()
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

func (p *SdiagProvider) FetchIfStale(since time.Duration) (err error) {
	if time.Since(p.LastUpdated()) > since {
		err = p.Fetch()
	}
	return err
}

// SdiagProvider data does not have a categorical filter, so this just returns the current data.
func (p *SdiagProvider) FilteredData() *TextData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()
	return &data
}
