package data

import "github.com/rivo/tview"


type NodeInfo struct {
	Name              string
	Partition         string
	State             string
	CPUs              string
	Memory            string
	CPULoad           string
	Reason            string
	ActiveFeatures    string
	AvailableFeatures string
	Sockets           string
	Cores             string
	Threads           string
	GRES              string
}

