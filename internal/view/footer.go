package view

import (
	"fmt"
	"time"

	"github.com/antvirf/stui/internal/config"
)

func (a *App) ShowNotification(text string, after time.Duration) {
	go func() {
		a.FooterMessage.SetText(text)
		time.Sleep(after)
		a.FooterMessage.Clear()
		a.App.Draw()
	}()
}

func (a *App) UpdateFooter(host, ip string) {
	// Left column
	a.FooterLineOne.SetText(
		fmt.Sprintf("Scheduler: %s (%s)", host, ip),
	)
	a.FooterLineTwo.SetText(
		fmt.Sprintf(
			"Data as of %s (%d ms)",
			a.LastUpdate.Format("15:04:05"),
			a.LastReqDuration.Milliseconds(),
		),
	)

	// Right column
	if a.NodesTableData != nil && config.NodeStatusField != "" {
		totalNodes := len(a.NodesTableData.Rows)
		// TODO: Compute number of unhealthy nodes or otherwise split by state
		a.FooterNodeStats.SetText(fmt.Sprintf("Nodes: %d", totalNodes))
	}

	if a.JobsTableData != nil && config.JobStatusField != "" {
		totalJobs := len(a.JobsTableData.Rows)
		// TODO: Compute number of running jobs or otherwise split by state
		a.FooterJobStats.SetText(fmt.Sprintf("Jobs: %d", totalJobs))
	}
}
