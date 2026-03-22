package view

const (
	STATE_KEY_SEARCH_FILTER = "SEARCH_FILTER"
)

func (a *App) LoadStateFromStateStore() {
	a.SearchPattern, _ = a.StateStore.GetStateKey(STATE_KEY_SEARCH_FILTER)
}

func (a *App) SaveState() {
	a.StateStore.SetStateKey(STATE_KEY_SEARCH_FILTER, a.SearchPattern)
	a.StateStore.SaveState()
}
