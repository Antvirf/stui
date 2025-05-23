package view

import (
	"time"

	"github.com/tiagomelo/go-clipboard/clipboard"
)

func (a *App) copyToClipBoard(text string) {
	c := clipboard.New()
	if err := c.CopyText(text); err != nil {
		a.ShowNotification(
			"[red]FAIL - clipboard not available, install libx11-dev / xorg-dev / libx11-devel [white]",
			2*time.Second,
		)
	} else {
		a.ShowNotification(
			"[green]Copied row details clipboard[white]",
			2*time.Second,
		)

	}

}
