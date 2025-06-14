package view

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) executeCommand(output *tview.TextView, cmdText string, pageName string) {
	output.SetText("\n\n")

	// Execute command
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
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

		// After a successful command...
		// ... clear the user's selection within the current view
		a.ClearSelectionFromCurrentView()

		// ... and trigger a table view refresh in the background
		a.RefreshAndRenderPage(pageName)

	}
}

func (a *App) ShowStandardCommandModal(command string, selectedMap map[string]bool, pageName string) {
	var selected []string
	for entry := range selectedMap {
		selected = append(selected, entry)
	}
	command = fmt.Sprintf("%s%s ", command, strings.Join(selected, ","))
	a.ShowCommandModal(command, pageName, false, false)
}

func (a *App) ShowCommandModal(command string, pageName string, executeImmediately bool, closeAfterExecute bool) {
	a.CommandModalOpen = true

	// Create input field with prefilled command
	input := tview.NewInputField().
		SetLabel("Command: ").
		SetFieldStyle(
			tcell.StyleDefault.Background(rowCursorColorBackground),
		).
		SetText(command).
		SetFieldWidth(0)

	if executeImmediately {
		input.SetDisabled(true)
	}

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

	// Create centered container with fixed size (80% width, 90% height)
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(modal, 0, 10, true),
			0, 16, false).
		AddItem(nil, 0, 1, false)

	// Store current page before showing modal
	previousFocus := a.App.GetFocus()

	// Add as overlay
	a.Pages.AddPage(COMMAND_PAGE, centered, true, true)
	a.App.SetFocus(input)

	if executeImmediately {
		a.executeCommand(output, input.GetText(), pageName)
		a.App.SetFocus(output)
		if closeAfterExecute {
			a.CloseCommandModal(COMMAND_PAGE, pageName, previousFocus)
			a.ShowNotification(
				fmt.Sprintf("[green]Executed command '%s'[white]", command),
				3*time.Second,
			)
			return
		}
	}

	// Set up input capture
	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			a.executeCommand(output, input.GetText(), pageName)
			return nil

		case tcell.KeyEsc:
			a.CloseCommandModal(COMMAND_PAGE, pageName, previousFocus)
			return nil
		}
		return event
	})

	// Set up output's input capture, so user can still escape out
	output.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.CloseCommandModal(COMMAND_PAGE, pageName, previousFocus)
			return nil
		}
		return event
	})
}

func (a *App) CloseCommandModal(commandPageName string, targetPage string, previousFocus tview.Primitive) {
	a.CommandModalOpen = false
	a.Pages.RemovePage(commandPageName)
	a.Pages.SwitchToPage(targetPage)
	a.App.SetFocus(previousFocus)
}
