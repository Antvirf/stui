package view

import "github.com/rivo/tview"

func init() {
	// Focused borders
	tview.Borders.HorizontalFocus = tview.BoxDrawingsHeavyHorizontal
	tview.Borders.VerticalFocus = tview.BoxDrawingsHeavyVertical
	tview.Borders.TopLeftFocus = tview.BoxDrawingsHeavyDownAndRight
	tview.Borders.TopRightFocus = tview.BoxDrawingsHeavyDownAndLeft
	tview.Borders.BottomLeftFocus = tview.BoxDrawingsHeavyUpAndRight
	tview.Borders.BottomRightFocus = tview.BoxDrawingsHeavyUpAndLeft

	// Non-focus borders
	tview.Borders.BottomLeft = tview.BoxDrawingsHeavyUpAndRight
	tview.Borders.BottomRight = tview.BoxDrawingsHeavyUpAndLeft
	tview.Borders.TopRight = tview.BoxDrawingsHeavyDownAndLeft
	tview.Borders.TopLeft = tview.BoxDrawingsHeavyDownAndRight
	tview.Borders.Horizontal = tview.BoxDrawingsHeavyHorizontal
	tview.Borders.Vertical = tview.BoxDrawingsHeavyVertical
}
