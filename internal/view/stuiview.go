package view

import (
	"fmt"
	"regexp"
	"slices"
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
	searchStringPointer *string,
) *StuiView {

	view := StuiView{
		titleHeader:                   title,
		provider:                      provider,
		Selection:                     make(map[string]bool),
		filter:                        "",
		searchEnabled:                 false,
		searchPattern:                 searchStringPointer,
		sortColumn:                    -1,        // No column sorted by default
		sortDirection:                 SORT_NONE, // Default to no sort
		updateTitleFunction:           updateTitleFunc,
		errorNotificationFunction:     errorNotifyFunc,
		dataStateNotificationFunction: dataStateNotifyFunc,
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

func (s *StuiView) Render() {
	startTime := time.Now()
	s.data = s.provider.FilteredData()
	filterDataTime := time.Since(startTime).Milliseconds()

	s.Table.Clear()

	// Compute counts
	totalCount := s.provider.Length()
	filteredCount := s.data.Length()

	searchFilterTime := int64(0)
	filteredRows := s.data.Rows
	if s.searchEnabled && *s.searchPattern != "" {
		filteredCount = 0 // Will be updated in the filtering loop below
		searchFilterStartTime := time.Now()
		filteredRows = [][]string{}

		pattern, err := regexp.Compile("(?i)" + *s.searchPattern)
		if err != nil {
			s.errorNotificationFunction(fmt.Sprintf("[red]Invalid search pattern: %v[white]", err))
		} else {
			// Preallocate slice with reasonable capacity
			filteredRows = make([][]string, 0, len(s.data.Rows)/2)

			for _, row := range s.data.Rows {
				// Check each column individually. We do NOT support entire-row search for performance reasons.
				matched := slices.ContainsFunc(row, func(cell string) bool { return pattern.MatchString(cell) })

				if matched {
					filteredRows = append(filteredRows, row)
					filteredCount++
				}
			}
		}
		searchFilterTime = time.Since(searchFilterStartTime).Milliseconds()
	}

	// Sort rows if sort column is set
	if s.sortColumn >= 0 && len(filteredRows) > 0 {
		sort.Slice(filteredRows, func(i, j int) bool {
			if s.sortDirection > 0 {
				return filteredRows[i][s.sortColumn] < filteredRows[j][s.sortColumn]
			}
			return filteredRows[i][s.sortColumn] > filteredRows[j][s.sortColumn]
		})
	}

	for col, header := range *s.data.Headers {
		// If header is a divided type, clean it up
		headerName := header.DisplayName

		// Add sort indicator if this is the sorted column
		if col == s.sortColumn {
			if s.sortDirection > 0 {
				headerName += " ↑"
			} else {
				headerName += " ↓"
			}
		}

		// Pad header with spaces to maintain width
		cell := tview.NewTableCell(headerName).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetTextColor(generalTextColor).
			SetAttributes(tcell.AttrBold).
			SetMaxWidth(len(header.DisplayName))

		// Highlight sorted column header
		if col == s.sortColumn {
			cell.SetBackgroundColor(selectionColor)
		} else {
			cell.SetBackgroundColor(generalBackgroundColor)
		}

		s.Table.SetCell(0, col, cell)
	}

	// Row and cell-level processing: Text wrapping, colorization, etc.
	for row, rowData := range filteredRows {
		var shouldColorizeRow bool

		// Check whether we should give this row a special color based on its state field
		colorizedColor, shouldColorizeRow := GetStateColorMapping(rowData[config.NodeViewColumnsStateIndex])

		for col, cell := range rowData {
			//logger.Debugf(fmt.Sprintf("'%-*s'", (*s.data.Headers)[col].Width, cell))
			// Op 1: Text wrapping
			colObject := (*s.data.Headers)[col]
			// We need to *pad* the text here, as tview does not support a 'minimum width' parameter for tables.
			cellView := tview.NewTableCell(fmt.Sprintf("%-*s", colObject.Width, cell)).
				SetAlign(tview.AlignLeft).
				SetExpansion(1)

			if colObject.FullWidthColumn {
				cellView.SetMaxWidth(0)
			} else {
				cellView.SetMaxWidth(colObject.Width)
			}

			// Highlight selected rows, or set color based on status
			if s.Selection[rowData[0]] {
				cellView.SetBackgroundColor(selectionColor)
				cellView.SetTextColor(selectionTextColor)
				cellView.SetSelectedStyle(tcell.StyleDefault.Background(selectionHighlightColor))
			} else {
				// Colorize text based on status
				if shouldColorizeRow {
					cellView.SetTextColor(colorizedColor)
				}

				// Other defaults
				cellView.SetBackgroundColor(generalBackgroundColor) // Explicitly set default when not selected
				cellView.SetSelectedStyle(tcell.StyleDefault.Background(rowCursorColorBackground))
			}

			s.Table.SetCell(row+1, col, cellView)
		}
	}

	// If no rows, set empty cells with spaces to maintain a nice looking column structure
	if len(filteredRows) == 0 {
		for col := range *s.data.Headers {
			spaces := strings.Repeat(" ", 1)
			s.Table.SetCell(1, col, tview.NewTableCell(spaces).
				SetAlign(tview.AlignLeft).
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
	s.updateTitleFunction(s.completeTitle)

	lastUpdated := s.provider.LastUpdated()
	timeSince := int(time.Since(lastUpdated).Seconds())
	s.dataStateNotificationFunction(fmt.Sprintf(
		"%s data as of %s (since %d seconds ago)",
		s.titleHeader,
		lastUpdated.Local().Format("15:04:05"),
		timeSince,
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

func GetStateColorMapping(text string) (color tcell.Color, hasMapping bool) {
	hasMapping = false
	color = tcell.ColorWhite

	// Process high priority mapping first
	for state, mappedColor := range STATE_COLORS_MAP_HIGH_PRIORITY {
		// We check using contains, as some states won't be an exact text match.
		// E.g. `CANCELLED` is sometimes `CANCELLED BY $UID`
		// E.g. `IDLE+DRAIN` is a valid node state, and should be interpreted as `DRAIN` for coloring.
		if strings.Contains(text, state) {
			color = mappedColor
			hasMapping = true
			return
		}
	}

	// Lower priority second
	for state, mappedColor := range STATE_COLORS_MAP {
		if strings.Contains(text, state) {
			color = mappedColor
			hasMapping = true
			return
		}
	}
	return
}
