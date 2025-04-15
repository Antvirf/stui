package view

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

func (a *App) ShowNotification(text string, after time.Duration) {
	go func() {
		a.FooterMessage.SetText(text)
		time.Sleep(after)
		a.FooterMessage.Clear()
		a.App.Draw()
	}()
}

func (a *App) UpdateHeader(hostNameWithIP string, lastRender time.Time, lastRenderDuration time.Duration) {
	// Left column
	a.HeaderLineOne.SetText(
		fmt.Sprintf("Scheduler: %s", hostNameWithIP),
	)
	a.HeaderLineTwo.SetText(
		fmt.Sprintf(
			"Last render: %s (%d ms)",
			lastRender.Format("15:04:05"),
			lastRenderDuration.Milliseconds(),
		),
	)
}

// Add each given primitive as a row to the top-left header area.
func (a *App) SetHeaderGridInnerContents(content ...tview.Primitive) {
	for index, entry := range content {
		a.HeaderGridInnerContents.AddItem(entry, index, FRST_COL, 1, 1, 1, 0, false)
	}
}
