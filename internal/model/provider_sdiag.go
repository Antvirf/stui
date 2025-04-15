package model

import (
	"time"

	"github.com/antvirf/stui/internal/config"
)

type SdiagProvider struct {
	BaseProvider[*TextData]
}

func NewSdiagProvider() *SdiagProvider {
	p := SdiagProvider{
		BaseProvider: NewBaseProvider[*TextData](),
	}
	p.Fetch()
	return &p
}

func (p *SdiagProvider) RunPeriodicRefresh(
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

func (p *SdiagProvider) Fetch() error {
	// TODO: Why does this deadlock?
	// p.mu.Lock()
	// defer p.mu.Unlock()

	rawData, err := getSdiagWithTimeout(config.RequestTimeout)

	if err != nil {
		p.updateError(err)
		return err
	}

	p.lastUpdated = time.Now()
	p.fetchCount++

	p.updateData(&TextData{Data: rawData})
	return nil
}

// SdiagProvider data does not have a categorical filter, so this just returns the current data.
func (p *SdiagProvider) FilteredData(filter string) *TextData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()
	return &data
}
