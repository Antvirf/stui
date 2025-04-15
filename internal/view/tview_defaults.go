package view

import "github.com/rivo/tview"

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
