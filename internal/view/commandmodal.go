package view

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) ShowCommandModal(commandFilter string, selectedMap map[string]bool) {
	a.CommandModalOpen = true
	var selected []string
	prefix := fmt.Sprintf("scontrol update %s=", commandFilter)
	for entry := range selectedMap {
		selected = append(selected, entry)
	}

	// Create input field with prefilled command
	input := tview.NewInputField().
		SetLabel("Command: ").
		SetFieldStyle(
			tcell.StyleDefault.Background(rowCursorColorBackground),
		).
		SetText(prefix + strings.Join(selected, ",") + " ").
		SetFieldWidth(0)

	// Create output view
	output := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	// Create flex layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true).
		AddItem(output, 0, 1, false)

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(" Execute Command (ESC to cancel) "),
			1, 0, false).
		AddItem(flex, 0, 1, true)

	modal.SetBorder(true).
		SetBorderColor(modalBorderColor).
		SetBackgroundColor(generalBackgroundColor)

	// Center the modal
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(modal, 25, 1, true),
			0, 60, false).
		AddItem(nil, 0, 1, false)

	// Store current page before showing modal
	previousPageName, _ := a.Pages.GetFrontPage()
	previousFocus := a.App.GetFocus()

	// Add as overlay
	pageName := "commandModal"
	a.Pages.AddPage(pageName, centered, true, true)
	a.App.SetFocus(input)

	// Set up input capture
	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			cmdText := input.GetText()
			output.SetText("Executing: " + cmdText + "\n\n")

			// Execute command
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "bash", "-c", cmdText)
			cmdOut, err := cmd.CombinedOutput()
			if err != nil {
				output.SetText(output.GetText(true) + "Error: " + err.Error() + "\n")
				output.SetText(output.GetText(true) + string(cmdOut))
			} else {
				commandOutput := string(cmdOut)
				if commandOutput == "" {
					output.SetText(output.GetText(true) + "Command executed successfully (no output)")

				} else {
					output.SetText(output.GetText(true) + commandOutput)
				}
				// Trigger table refresh in the background after a successful command
				go a.UpdateAllViews()
			}

			return nil

		case tcell.KeyEsc:
			a.CommandModalOpen = false
			a.Pages.RemovePage(pageName)
			a.Pages.SwitchToPage(previousPageName)
			a.App.SetFocus(previousFocus)
			return nil
		}
		return event
	})
}
