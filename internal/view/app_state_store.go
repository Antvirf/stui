package view

const (
	STATE_KEY_SEARCH_FILTER = "SEARCH_FILTER"
)

func (a *App) LoadStateFromStateStore() {
	a.SearchPattern = a.StateStore.State.SearchFilter
}

func (a *App) SaveState() {
	a.StateStore.State.SearchFilter = a.SearchPattern
	_, a.StateStore.State.PartitionFilter = a.PartitionSelector.GetCurrentOption()
	_, a.StateStore.State.NodesPane.StateFilter = a.NodeStateSelector.GetCurrentOption()
	_, a.StateStore.State.JobsPane.StateFilter = a.JobStateSelector.GetCurrentOption()
	_, a.StateStore.State.SacctmgrPane.EntityFilter = a.SacctMgrEntitySelector.GetCurrentOption()

	// TODO: Store state of sort columns and directions
	//a.StateStore.State.NodesPane.SortByColumn = a.NodesView.sortColumn
	//a.StateStore.State.NodesPane.SortDirection = a.NodesView.sortDirection
	//a.StateStore.State.JobsPane.SortByColumn = a.JobsView.sortColumn
	//a.StateStore.State.JobsPane.SortDirection = a.JobsView.sortDirection
	//a.StateStore.State.SacctPane.SortByColumn = a.SacctView.sortColumn
	//a.StateStore.State.SacctPane.SortDirection = a.SacctView.sortDirection
	//a.StateStore.State.SacctPane.SortByColumn = a.SacctView.sortColumn
	//a.StateStore.State.SacctPane.SortDirection = a.SacctView.sortDirection
	//a.StateStore.State.SacctmgrPane.SortByColumn = a.SacctMgrView.sortColumn
	//a.StateStore.State.SacctmgrPane.SortDirection = a.SacctMgrView.sortDirection
	//a.StateStore.State.SacctmgrPane.SortByColumn = a.SacctMgrView.sortColumn
	//a.StateStore.State.SacctmgrPane.SortDirection = a.SacctMgrView.sortDirection
	a.StateStore.SaveState()
}
