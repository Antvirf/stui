package view

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/antvirf/stui/internal/config"
	"github.com/antvirf/stui/internal/logger"
	"github.com/antvirf/stui/internal/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewStuiView(
	title string,
	provider model.DataProvider[*model.TableData],
	updateTitleFunc func(string) *tview.Box,
	errorNotifyFunc func(string),
	dataStateNotifyFunc func(string),
	cellClickFunction func(string),
	headerClickFunction func(int) *tview.DropDown,
	searchStringPointer *string,
) *StuiView {

	view := StuiView{
		titleHeader:                   title,
		provider:                      provider,
		Selection:                     make(map[string]bool), // The actual true/false is NOT used, only map presence.
		filter:                        "",
		searchEnabled:                 false,
		searchPattern:                 searchStringPointer,
		sortColumn:                    -1,        // No column sorted by default
		sortDirection:                 SORT_NONE, // Default to no sort
		updateTitleFunction:           updateTitleFunc,
		errorNotificationFunction:     errorNotifyFunc,
		dataStateNotificationFunction: dataStateNotifyFunc,
		cellClickFunction:             cellClickFunction,
		headerClickFunction:           headerClickFunction,
	}

	view.Table = tview.NewTable()
	view.Table.
		SetBorders(false). // Remove all borders
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	view.Table.
		SetEvaluateAllRows(false). // Do not evalute all rows when rendering.
		SetFixed(1, 0).            // Fixed header row
		SetSelectable(true, false) // Selectable rows but not columns
	view.Table.SetSelectedStyle(tcell.StyleDefault.
		Background(rowCursorColorBackground).
		Foreground(rowCursorColorForeground))
	view.Table.SetBackgroundColor(generalBackgroundColor)
	view.Grid = tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		AddItem(view.Table, 0, 0, 1, 1, 0, 0, true)
	return &view
}

type StuiViewInt interface {
	// Renders this component in tview, without affecting underlying data
	Render()

	// Sets regex filter for rows filtering
	SetFilter()

	// Sets search status
	SetSearchEnabled(bool)

	// Updates data from provider and renders the view
	FetchAndRender()
}

type StuiView struct {
	// View components
	Table         *tview.Table
	Grid          *tview.Grid
	Selection     map[string]bool
	titleHeader   string
	completeTitle string
	searchEnabled bool
	searchPattern *string // Pointer to a shared string

	// Sorting state
	sortColumn    int // Index of column being sorted (-1 for none)
	sortDirection int // -1 for descending, 1 for ascending

	// Callback functions
	updateTitleFunction           func(string) *tview.Box
	errorNotificationFunction     func(string)
	dataStateNotificationFunction func(string)
	cellClickFunction             func(string)
	headerClickFunction           func(int) *tview.DropDown

	// Data components
	provider model.DataProvider[*model.TableData]
	data     *model.TableData
	filter   string
}

func (s *StuiView) SetFilter(filter string) {
	s.filter = filter
}

func (s *StuiView) SetTitleHeader(v string) {
	s.titleHeader = v
}

func (s *StuiView) SetSearchEnabled(value bool) {
	s.searchEnabled = value
}

// Apply regex filter to current data and return only the matching rows
func (s *StuiView) ApplyRegexSearchFilterToRows() ([][]model.CellValue, int, int64) {
	searchFilterStartTime := time.Now()
	filteredCount := s.data.Length()
	filteredRows := s.data.Rows
	if s.searchEnabled && *s.searchPattern != "" {
		if re, err := regexp.Compile("(?i)" + *s.searchPattern); err == nil {
			filteredRows = make([][]model.CellValue, 0, len(s.data.Rows)/2)
			for i, row := range s.data.Rows {
				if re.MatchString(s.data.RowsAsSingleStrings[i]) {
					filteredRows = append(filteredRows, row)
					filteredCount++
				}
			}
		} else {
			s.errorNotificationFunction(fmt.Sprintf("[red]Invalid search pattern: %v[white]", err))
		}
	}
	return filteredRows, filteredCount, time.Since(searchFilterStartTime).Milliseconds()
}

func (s *StuiView) ApplySortingToRows(rows *[][]model.CellValue) {
	if s.sortColumn >= 0 && len(*rows) > 0 {
		sort.Slice(*rows, func(i, j int) bool {
			cmp := (*rows)[i][s.sortColumn].Compare((*rows)[j][s.sortColumn])
			if s.sortDirection > 0 {
				return cmp < 0
			}
			return cmp > 0
		})
	}
}

func (s *StuiView) GetHeadersAsText() []string {
	rawHeaders := *s.data.Headers
	headers := make([]string, len(rawHeaders))
	for i, h := range rawHeaders {
		headers[i] = h.DisplayName
	}
	return headers
}

// GetVisibleRowsAsText returns the currently visible rows (post-filter, search, sort).
func (s *StuiView) GetVisibleRowsAsText() (rows [][]string) {
	if s.data == nil {
		return nil
	}

	filteredRows, _, _ := s.ApplyRegexSearchFilterToRows()
	s.ApplySortingToRows(&filteredRows)

	// .Display() for each cell to get nice formatting
	rows = make([][]string, len(filteredRows))
	for i, row := range filteredRows {
		r := make([]string, len(row))
		for j, cell := range row {
			r[j] = cell.Display()
		}
		rows[i] = r
	}
	return
}

func (s *StuiView) Render() {
	startTime := time.Now()
	s.data = s.provider.FilteredData()
	filterDataTime := time.Since(startTime).Milliseconds()

	s.Table.Clear()

	totalCount := s.provider.Length()
	searchFilterTime := int64(0)
	filteredRows, filteredCount, searchFilterTime := s.ApplyRegexSearchFilterToRows()
	s.ApplySortingToRows(&filteredRows)

	for col, header := range *s.data.Headers {
		// If header is a divided type, clean it up
		headerName := header.DisplayName

		// Add sort indicator if this is the sorted column
		if col == s.sortColumn {
			if s.sortDirection > 0 {
				headerName = "↑ " + headerName
			} else {
				headerName = "↓ " + headerName
			}
		}

		// Pad header with spaces to maintain width
		cell := tview.NewTableCell(headerName).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetTextColor(generalTextColor).
			SetAttributes(tcell.AttrBold).
			SetMaxWidth(len(header.DisplayName))

		// Sort on click
		cell.SetClickedFunc(func() bool {
			s.headerClickFunction(col + 1)
			return true
		})

		// Highlight sorted column header
		if col == s.sortColumn {
			cell.SetBackgroundColor(selectionColor)
		} else {
			cell.SetBackgroundColor(generalBackgroundColor)
		}

		s.Table.SetCell(0, col, cell)
	}

	// TODO: Move the regex compilation to searchbox.go
	// and add the app/stuiview links as pointers to the regex instead of string
	var searchPatternRegex *regexp.Regexp
	searchPatternRegex, _ = regexp.Compile("(?i)" + *s.searchPattern)

	// Row and cell-level processing: Text wrapping, colorization, etc.
	for row, rowData := range filteredRows {
		var colorizedColor tcell.Color
		var shouldColorizeRow bool

		// Check whether we should give this row a special color based on its state field
		if len(rowData) > config.NodeViewColumnsStateIndex {
			colorizedColor, shouldColorizeRow = GetStateColorMapping(rowData[config.NodeViewColumnsStateIndex].Display())
		} else {
			colorizedColor, shouldColorizeRow = generalBackgroundColor, false
		}

		// Pre-calculate row-level search matches (used for highlights) for efficiency
		var matches [][]int
		if s.searchEnabled && *s.searchPattern != "" && !config.DisableSearchHighlight {
			var sb strings.Builder
			for _, c := range rowData {
				sb.WriteString(c.Display())
			}
			matches = searchPatternRegex.FindAllStringIndex(sb.String(), -1)
		}

		currentOffset := 0
		matchIdx := 0
		for col, cell := range rowData {
			// Op 1: Text wrapping
			colObject := (*s.data.Headers)[col]
			cellText := cell.Display()
			cellView := tview.NewTableCell(fmt.Sprintf("%-*s", colObject.Width, cellText)).
				SetAlign(tview.AlignLeft).
				SetExpansion(1)

			// Op 2: Highlight search results
			isMatched := false
			cellEnd := currentOffset + len(cellText)
			for i := matchIdx; i < len(matches); i++ {
				m := matches[i]
				if m[1] <= currentOffset {
					matchIdx = i + 1
					continue
				}
				if m[0] >= cellEnd {
					break
				}
				isMatched = true
				break
			}

			cellView.SetClickedFunc(func() bool {
				s.cellClickFunction(cellText)
				return true
			})

			if colObject.FullWidthColumn {
				cellView.SetMaxWidth(0)
			} else {
				cellView.SetMaxWidth(colObject.Width)
			}

			_, isSelected := s.Selection[rowData[0].Display()]
			colorizeTableCell(cellView, isSelected, isMatched, shouldColorizeRow, colorizedColor)

			s.Table.SetCell(row+1, col, cellView)
			currentOffset = cellEnd
		}
	}

	// If no rows, set empty cells with spaces to maintain a nice looking column structure
	if len(filteredRows) == 0 {
		for col := range *s.data.Headers {
			spaces := strings.Repeat(" ", 1)
			s.Table.SetCell(1, col, tview.NewTableCell(spaces).
				SetAlign(tview.AlignLeft).
				SetSelectable(false).
				SetMaxWidth(0).
				SetExpansion(1))
		}
	}

	// Callbacks
	s.completeTitle = fmt.Sprintf(
		" %s ( %s/%s ) ",
		s.titleHeader,
		FormatNumberWithCommas(filteredCount),
		FormatNumberWithCommas(totalCount),
	)
	selectionCount := getSelectionCount(&s.Selection)
	if selectionCount > 0 {
		s.completeTitle += fmt.Sprintf("| %s selected", FormatNumberWithCommas(selectionCount))
	}
	s.updateTitleFunction(s.completeTitle)

	lastUpdated := s.provider.LastUpdated()
	s.dataStateNotificationFunction(fmt.Sprintf(
		"%s data as of %s",
		s.titleHeader,
		lastUpdated.Local().Format("15:04:05"),
	))

	if s.provider.LastError() != nil {
		s.errorNotificationFunction(fmt.Sprintf(
			"[red]%s [white]", s.provider.LastError(),
		))
	} else {
		s.errorNotificationFunction("")
	}

	execTime := time.Since(startTime).Milliseconds()
	searchInfo := ""
	if s.searchEnabled {
		searchInfo = fmt.Sprintf(", search_filter=%dms", searchFilterTime)
	}
	logger.Debugf("%s: render completed in %dms (filter_data_time=%dms%s, rows=%d)",
		s.titleHeader, execTime, filterDataTime, searchInfo, filteredCount)
}

func (s *StuiView) FetchIfStaleAndRender(since time.Duration) {
	if time.Since(s.provider.LastUpdated()) > since {
		s.FetchAndRender()
	} else {
		s.Render()
	}
}

func (s *StuiView) FetchAndRender() {
	s.provider.Fetch()
	s.Render()
}

// colorizeTableCell applies consistent color and style to a table cell based on its state
func colorizeTableCell(
	cell *tview.TableCell,
	isSelected bool,
	isSearchMatched bool,
	shouldColorizeRow bool,
	colorizedColor tcell.Color,
) {
	// 1. Determine base colors based on status and selection
	targetTextColor := generalTextColor
	targetBgColor := generalBackgroundColor

	if shouldColorizeRow {
		targetTextColor = colorizedColor
	}

	if isSelected {
		targetBgColor = selectionColor
		if !shouldColorizeRow {
			targetTextColor = selectionTextColor
		}
	}

	// 2. Search highlights take precedence over normal text colors if matched
	if isSearchMatched {
		targetTextColor = searchHighlightFgColor
		targetBgColor = searchHighlightBgColor
	}

	cell.SetTextColor(targetTextColor).SetBackgroundColor(targetBgColor)
	cell.SetReference(isSearchMatched)
	cell.SetAttributes(tcell.AttrNone)

	// 3. Define selected style (cursor appearance)
	cursorBgColor := rowCursorColorBackground
	if isSelected {
		cursorBgColor = selectionHighlightColor
	}

	// We want to keep the text color when highlighted if it's a state color
	cursorTextColor := targetTextColor
	if isSearchMatched {
		cursorTextColor = searchHighlightFgColor
		cursorBgColor = searchHighlightHoverBgColor
	}

	cell.SetSelectedStyle(tcell.StyleDefault.
		Background(cursorBgColor).
		Foreground(cursorTextColor))
}

func GetStateColorMapping(text string) (color tcell.Color, hasMapping bool) {
	hasMapping = false
	color = tcell.ColorWhite

	// Process priority list. Earlier entries take precedence.
	for _, pattern := range StatePatterns {
		// We use word boundaries to ensure we don't match substrings.
		// E.g. we don't want "DOWN" to match "POWERED_DOWN".
		// Note: Slurm states can be combined with "+", which acts as a word boundary.
		if pattern.Regexp.MatchString(text) {
			color = pattern.Color
			hasMapping = true
			return
		}
	}
	return
}
