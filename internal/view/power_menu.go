package view

import (
	"fmt"
	"slices"
	"strings"

	"github.com/antvirf/stui/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var nodePowerStates = []string{"POWER_UP", "POWER_DOWN"}

func nodePowerStateIndex(shortcut rune) int {
	switch shortcut {
	case 'u', 'U':
		return 0
	case 'd', 'D':
		return 1
	default:
		return -1
	}
}

func nodePowerCommand(nodes map[string]bool, state string) string {
	nodeNames := make([]string, 0, len(nodes))
	for nodeName := range nodes {
		nodeNames = append(nodeNames, nodeName)
	}
	slices.Sort(nodeNames)
	return fmt.Sprintf("scontrol update NodeName=%q State=%q", strings.Join(nodeNames, ","), state)
}

func (a *App) ShowNodePowerMenu(nodes map[string]bool) {
	if len(nodes) == 0 {
		return
	}

	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	for _, state := range nodePowerStates {
		state := state
		shortcut := []rune(strings.ToLower(state))[len("power_")]
		list.AddItem(fmt.Sprintf("(%c) %s", shortcut, state), "", 0, func() {
			a.Pages.RemovePage("node-power-menu")
			a.ShowCommandModal(
				nodePowerCommand(nodes, state),
				config.NODES_PAGE,
				false,
				false,
			)
		})
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
			SetText(" Node Power (U/D to select, Enter to confirm, ESC to cancel) "),
			1, 0, false).
		AddItem(list, 0, 1, true)
	modal.SetBorder(true).
		SetBorderColor(modalBorderColor).
		SetBackgroundColor(generalBackgroundColor)

	verticallyCentered := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(modal, 7, 0, true).
		AddItem(nil, 0, 1, false)
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(verticallyCentered, 46, 0, true).
		AddItem(nil, 0, 1, false)

	a.Pages.AddPage("node-power-menu", centered, true, true)
	a.App.SetFocus(list)
}
