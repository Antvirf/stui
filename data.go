package main

import "github.com/rivo/tview"

type TableData struct {
	Headers []string
	Rows    [][]string
}

type NodeInfo struct {
	Name          string
	Partition     string
	State         string
	CPUs          string
	Memory        string
	CPULoad       string
	Reason        string
	ActiveFeatures string
	AvailableFeatures string
	Sockets       string
	Cores         string
	Threads       string
	GRES          string
}

func renderTable(table *tview.Table, data TableData) {
	table.Clear()
	
	// Set headers with fixed width
	columnWidths := []int{10, 10, 10, 6, 8, 8, 20, 6, 6, 6, 15} // Adjust as needed
	for col, header := range data.Headers {
		table.SetCell(0, col, tview.NewTableCell(header).
			SetSelectable(false).
			SetAlign(tview.AlignCenter).
			SetMaxWidth(columnWidths[col])
	}

	// Set rows with text wrapping
	for row, rowData := range data.Rows {
		for col, cell := range rowData {
			table.SetCell(row+1, col, tview.NewTableCell(cell).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(columnWidths[col]).
				SetExpansion(1)
		}
	}
}
