package view

import (
	"fmt"
	"regexp"
	"strings"
	"time"

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
		updateTitleFunction:           updateTitleFunc,
		errorNotificationFunction:     errorNotifyFunc,
		dataStateNotificationFunction: dataStateNotifyFunc,
	}

	view.Table = tview.NewTable()
	view.Table.
		SetBorders(false). // Remove all borders
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 1, 1, 1) // Top, right, bottom, left padding
	view.Table.SetFixed(1, 0)             // Fixed header row
	view.Table.SetSelectable(true, false) // Selectable rows but not columns
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
	searchEnabled bool
	searchPattern *string // Pointer to a shared string

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

		pattern, err := regexp.Compile("(?i)" + regexp.QuoteMeta(*s.searchPattern))
		if err != nil {
			s.errorNotificationFunction(fmt.Sprintf("[red]Invalid search pattern: %v[white]", err))
		} else {
			// Preallocate slice with reasonable capacity
			filteredRows = make([][]string, 0, len(s.data.Rows)/2)

			for _, row := range s.data.Rows {
				// Check each column individually to avoid string concatenation when possible
				matched := false
				for _, cell := range row {
					if pattern.MatchString(cell) {
						matched = true
						break
					}
				}

				if matched {
					filteredRows = append(filteredRows, row)
					filteredCount++
				}
			}
		}
		searchFilterTime = time.Since(searchFilterStartTime).Milliseconds()
	}

	for col, header := range *s.data.Headers {

		// If header is a divided type, clean it up
		headerName := header.Name
		if header.DividedByColumn {
			headerName = strings.Replace(header.Name, "//", "/", -1)
		}

		// Pad header with spaces to maintain width
		s.Table.SetCell(0, col, tview.NewTableCell(headerName).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetBackgroundColor(generalBackgroundColor).
			SetTextColor(generalTextColor).
			SetAttributes(tcell.AttrBold))
	}

	// Set rows with text wrapping
	for row, rowData := range filteredRows {
		for col, cell := range rowData {
			cellView := tview.NewTableCell(cell).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(0).
				SetExpansion(1)

			// Highlight selected rows
			if s.Selection[rowData[0]] {
				cellView.SetBackgroundColor(selectionColor)
			} else {
				cellView.SetBackgroundColor(generalBackgroundColor) // Explicitly set default when not selected
			}

			s.Table.SetCell(row+1, col, cellView)
		}
	}

	// If no rows, set empty cells with spaces to maintain a nice looking column structure
	if len(filteredRows) == 0 {
		for col, _ := range *s.data.Headers {
			spaces := strings.Repeat(" ", 1)
			s.Table.SetCell(1, col, tview.NewTableCell(spaces).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(0).
				SetExpansion(1))
		}
	}

	// Callbacks
	s.updateTitleFunction(fmt.Sprintf(
		" %s (%s / %s) ", s.titleHeader, FormatNumberWithCommas(filteredCount), FormatNumberWithCommas(totalCount),
	))

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
