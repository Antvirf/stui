package view

import (
	"fmt"
	"time"

	"github.com/tiagomelo/go-clipboard/clipboard"
)

func (a *App) copyCellToClipBoard(text string) {
	a.copyToClipBoard(text, fmt.Sprintf("[green]Copied cell text: %s[white]", text))
}

func (a *App) copyToClipBoard(text string, success string) {
	c := clipboard.New()
	if err := c.CopyText(text); err != nil {
		a.ShowNotification(
			"[red]FAIL - clipboard not available, install libx11-dev / xorg-dev / libx11-devel [white]",
			2*time.Second,
		)
	} else {
		a.ShowNotification(success, 2*time.Second)
	}

}
