package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	// These constants make tview code more readable in column/row assignment operations.
	FRST_ROW = 0
	SCND_ROW = 1
	THRD_ROW = 2
	FRTH_ROW = 3
	FFTH_ROW = 4
	FRST_COL = 0
	SCND_COL = 1
	THRD_COL = 2
	FRTH_COL = 3
	FFTH_COL = 4

	// Color configs for the app
	generalBackgroundColor     = tcell.ColorBlack
	generalTextColor           = tcell.Color255
	pagesBorderColor           = tcell.Color240 // Medium gray
	rowCursorColorBackground   = tcell.Color240
	rowCursorColorForeground   = tcell.ColorWhite
	selectionTextColor         = tcell.ColorWhite
	selectionColor             = tcell.ColorCadetBlue
	selectionHighlightColor    = tcell.ColorDeepSkyBlue
	paneSelectorHighlightColor = tcell.Color172 // Orange
	modalBorderColor           = tcell.Color172 // Orange
	dropdownBackgroundColor    = tcell.Color240 // Medium gray
	dropdownForegroundColor    = tcell.Color255 // White
	searchboxLabelColor        = tcell.ColorOrange
)

func init() {
	// Focused borders
	tview.Borders.HorizontalFocus = tview.BoxDrawingsLightHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsLightVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsLightDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsLightUpAndLeft

	// Non-focus borders
	tview.Borders.BottomLeft = tview.BoxDrawingsLightUpAndRight
	tview.Borders.BottomRight = tview.BoxDrawingsLightUpAndLeft
	tview.Borders.TopRight = tview.BoxDrawingsLightDownAndLeft
	tview.Borders.TopLeft = tview.BoxDrawingsLightDownAndRight
	tview.Borders.Horizontal = tview.BoxDrawingsLightHorizontal
	tview.Borders.Vertical = tview.BoxDrawingsLightVertical
}
