package model

import (
	"fmt"
	"strings"

	"github.com/antvirf/stui/internal/config"
)

type SacctMgrProvider struct {
	BaseProvider[*TableData]
}

func NewSacctMgrProvider() *SacctMgrProvider {
	p := SacctMgrProvider{
		BaseProvider: BaseProvider[*TableData]{},
	}
	p.Fetch()
	return &p
}

func (p *SacctMgrProvider) Fetch() error {
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
		emptyData := &TableData{
			Headers: &[]config.ColumnConfig{},
			Rows:    [][]string{},
		}
		p.updateData(emptyData)
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
	return nil
}

// SacctMgrProvider data does not have any categorical filters, so this just returns the current data.
func (p *SacctMgrProvider) FilteredData() *TableData {
	return p.data
}
