package view

import (
	"time"
)

func (a *App) ShowNotification(text string, after time.Duration) {
	go func() {
		a.FooterMessage.SetText(text)
		time.Sleep(after)
		a.FooterMessage.Clear()
		a.App.Draw()
	}()
}
