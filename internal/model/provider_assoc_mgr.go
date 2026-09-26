package model

import (
	"time"

	"github.com/antvirf/stui/internal/config"
)

var assocMgrColumns = []config.ColumnConfig{
	{RawName: "ClusterName", DisplayName: "Cluster"},
	{RawName: "Account", DisplayName: "Account"},
	{RawName: "UserName", DisplayName: "User"},
	{RawName: "Partition", DisplayName: "Partition"},
	{RawName: "Priority", DisplayName: "Priority"},
	{RawName: "GrpJobs", DisplayName: "GrpJobs"},
	{RawName: "GrpJobsAccrue", DisplayName: "GrpAccrue"},
	{RawName: "GrpSubmitJobs", DisplayName: "GrpSubmit"},
	{RawName: "GrpWall", DisplayName: "GrpWall"},
	{RawName: "GrpTRES", DisplayName: "GrpTRES"},
	{RawName: "GrpTRESMins", DisplayName: "GrpTRESMins"},
	{RawName: "GrpTRESRunMins", DisplayName: "GrpTRESRunMins"},
	{RawName: "MaxJobs", DisplayName: "MaxJobs"},
	{RawName: "MaxJobsAccrue", DisplayName: "MaxAccrue"},
	{RawName: "MaxSubmitJobs", DisplayName: "MaxSubmit"},
	{RawName: "MaxWallPJ", DisplayName: "MaxWallPJ"},
	{RawName: "MaxTRESPJ", DisplayName: "MaxTRESPJ"},
	{RawName: "MaxTRESPN", DisplayName: "MaxTRESPN"},
	{RawName: "MaxTRESMinsPJ", DisplayName: "MaxTRESMinsPJ"},
}

type AssocMgrProvider struct {
	BaseProvider[*TableData]
}

func NewAssocMgrProvider(loadData bool) *AssocMgrProvider {
	columns := append([]config.ColumnConfig(nil), assocMgrColumns...)
	p := AssocMgrProvider{
		BaseProvider: BaseProvider[*TableData]{
			data: EmptyTableDataWithColumns(&columns),
		},
	}
	if loadData {
		p.Fetch()
	}
	return &p
}

func (p *AssocMgrProvider) Fetch() error {
	computeColumnWidths := p.lastUpdated.IsZero()
	rawData, err := getScontrolDataWithTimeout(
		"show assoc_mgr flags=assoc",
		p.data.Headers,
		config.RequestTimeout,
		computeColumnWidths,
		parseAssocMgrOutput,
	)
	if err != nil {
		p.updateError(err)
		return err
	}

	p.updateData(rawData)
	return nil
}

func (p *AssocMgrProvider) FetchIfStale(since time.Duration) (err error) {
	if time.Since(p.LastUpdated()) > since {
		err = p.Fetch()
	}
	return err
}

func (p *AssocMgrProvider) FilteredData() *TableData {
	return p.data
}
