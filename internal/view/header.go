package view

import (
	"time"

	"github.com/rivo/tview"
)

func (a *App) ShowNotification(text string, after time.Duration) {
	go func() {
		a.HeaderLineThree.SetText(text)
		time.Sleep(after)
		a.HeaderLineThree.Clear()
		a.App.Draw()
	}()
}

func (a *App) UpdateHeaderLineOne(v string) {
	a.HeaderLineOne.SetText(v)
}

func (a *App) UpdateHeaderLineTwo(v string) {
	a.HeaderLineTwo.SetText(v)
}

// Add each given primitive as a row to the top-left header area.
func (a *App) SetHeaderGridInnerContents(content ...tview.Primitive) {
	a.HeaderGridInnerContents.Clear()
	for index, entry := range content {
		a.HeaderGridInnerContents.AddItem(entry, index, FRST_COL, 1, 1, 1, 0, false)
	}
}
