package view

import (
	"time"

	"github.com/antvirf/stui/internal/config"
	"golang.design/x/clipboard"
)

func (a *App) copyToClipBoard(text string) {
	if config.ClipboardAvailable {
		clipboard.Write(clipboard.FmtText, []byte(text))
		a.ShowNotification(
			"[green]Copied row details clipboard[white]",
			2*time.Second,
		)
	} else {
		a.ShowNotification(
			"[red]FAIL - clipboard not available, install libx11-dev / xorg-dev / libx11-devel [white]",
			2*time.Second,
		)
	}
}
