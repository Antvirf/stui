package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
)

type SacctMgrProvider struct {
	BaseProvider[*TableData]
}

func NewSacctMgrProvider() *SacctMgrProvider {
	p := SacctMgrProvider{
		BaseProvider: NewBaseProvider[*TableData](),
	}
	p.Fetch()
	return &p
}

func (p *SacctMgrProvider) RunPeriodicRefresh(
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

func (p *SacctMgrProvider) Fetch() error {
	// TODO: Why does this deadlock?
	// p.mu.Lock()
	// defer p.mu.Unlock()

	var columns []config.ColumnConfig
	columnConfig := strings.Split(SACCTMGR_ENTITY_COLUMN_CONFIGS[config.SacctMgrCurrentEntity], ",")
	for _, key := range columnConfig {
		columns = append(columns, config.ColumnConfig{Name: key})
	}

	rawData, err := getSacctMgrDataWithTimeout(
		fmt.Sprintf("show %s --parsable2", config.SacctMgrCurrentEntity),
		config.RequestTimeout,
		&columns,
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

// SacctMgrProvider data does not have a categorical filter, so this just returns the current data.
func (p *SacctMgrProvider) FilteredData(filter string) *TableData {
	p.mu.RLock()
	defer p.mu.RUnlock()
	data := *p.data.DeepCopy()
	return &data
}
