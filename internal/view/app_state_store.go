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
	a.StateStore.SaveState()
}
