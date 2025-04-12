package view

import "github.com/rivo/tview"

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
