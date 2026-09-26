package view

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/antvirf/stui/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type nodePowerOption struct {
	command     string
	shortcut    rune
	description string
}

var nodePowerOptions = []nodePowerOption{
	{"power up", 'u', "Run ResumeProgram to leave power-saving mode."},
	{"power down", 'd', "Run SuspendProgram to enter power-saving mode."},
	{"power down asap", 'a', "Drain node; power down after running jobs finish."},
	{"power down force", 'f', "Cancel jobs, power down, then reset to IDLE."},
	{"reboot", 'r', "Reboot when the node is idle."},
	{"reboot asap", 'b', "Drain node; reboot after running jobs finish."},
	{"cancel_reboot", 'c', "Cancel a pending reboot request."},
}

func nodePowerStateIndex(shortcut rune) int {
	for index, option := range nodePowerOptions {
		if unicode.ToLower(shortcut) == option.shortcut {
			return index
		}
	}
	return -1
}

func nodePowerCommand(nodes map[string]bool, command string) string {
	nodeNames := make([]string, 0, len(nodes))
	for nodeName := range nodes {
		nodeNames = append(nodeNames, nodeName)
	}
	slices.Sort(nodeNames)
	return fmt.Sprintf("scontrol %s %q", command, strings.Join(nodeNames, ","))
}

func (a *App) ShowNodePowerMenu(nodes map[string]bool) {
	if len(nodes) == 0 {
		return
	}

	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	for _, option := range nodePowerOptions {
		option := option
		list.AddItem(
			fmt.Sprintf("(%c) %-16s %s", option.shortcut, option.command, option.description),
			"",
			0,
			func() {
				a.Pages.RemovePage("node-power-menu")
				a.ShowCommandModal(
					nodePowerCommand(nodes, option.command),
					config.NODES_PAGE,
					false,
					false,
				)
			},
		)
	}

	previousFocus := a.App.GetFocus()
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			a.Pages.RemovePage("node-power-menu")
			a.App.SetFocus(previousFocus)
			return nil
		}

		if index := nodePowerStateIndex(event.Rune()); index >= 0 {
			list.SetCurrentItem(index)
			return nil
		}
		return event
	})

	modal := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(" Node Power (U/D/A/F/R/B/C to select, Enter to confirm, ESC to cancel) "),
			1, 0, false).
		AddItem(list, 0, 1, true)
	modal.SetBorder(true).
		SetBorderColor(modalBorderColor).
		SetBackgroundColor(generalBackgroundColor)

	verticallyCentered := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(modal, 10, 0, true).
		AddItem(nil, 0, 1, false)
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(verticallyCentered, 86, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("node-power-menu", centered, true, true)
	a.App.SetFocus(list)
}
