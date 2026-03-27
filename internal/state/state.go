package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/antvirf/stui/internal/logger"
)

type StuiState struct {
	active    bool
	statePath string
	State     StuiStateStruct
}

type StuiStateStruct struct {
	SearchFilter    string            `json:"searchFilter"`
	PartitionFilter string            `json:"partitionFilter"`
	NodesPane       NodesPaneState    `json:"nodesPane"`
	JobsPane        JobsPaneState     `json:"jobsPane"`
	SacctPane       SacctPaneState    `json:"sacctPane"`
	SacctmgrPane    SacctmgrPaneState `json:"sacctmgrPane"`
}

type NodesPaneState struct {
	StateFilter string `json:"stateFilter"`
	//SortByColumn  int    `json:"sortByColumn"`
	//SortDirection int    `json:"sortDirection"`
}
type JobsPaneState struct {
	StateFilter string `json:"stateFilter"`
	//SortByColumn  int    `json:"sortByColumn"`
	//SortDirection int    `json:"sortDirection"`
}
type SacctPaneState struct {
	//SortByColumn  int `json:"sortByColumn"`
	//SortDirection int `json:"sortDirection"`
}
type SacctmgrPaneState struct {
	EntityFilter string `json:"entityFilter"`
	//SortByColumn  int    `json:"sortByColumn"`
	//SortDirection int    `json:"sortDirection"`
}

// Sets up state file and parent directories if they don't yet exist.
// Loads state from pre-existing files if available.
func InitializeStuiState(basePath string) (*StuiState, error) {
	StateStore := &StuiState{
		statePath: path.Join(basePath, "stui-state.json"),
		State: StuiStateStruct{
			SearchFilter: "",
			NodesPane:    NodesPaneState{},
			JobsPane:     JobsPaneState{},
			SacctPane:    SacctPaneState{},
			SacctmgrPane: SacctmgrPaneState{},
		},
		active: basePath != "",
	}

	if !StateStore.active {
		logger.Debugf("state store: path is null, state not loaded")
		return StateStore, nil
	}

	logger.Debugf("state store: initializing with path: '%s'", StateStore.statePath)

	if _, err := os.Stat(StateStore.statePath); err != nil {
		logger.Debug("state store: target does not exist, will create")
		dir := filepath.Dir(StateStore.statePath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return StateStore, fmt.Errorf("failed to create state directory: %w", err)
			}
		}

		file, _ := os.Create(StateStore.statePath)
		emptyState := make(map[string]string)
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(emptyState); err != nil {
			return StateStore, fmt.Errorf("failed to write empty state: %w", err)
		}

		return StateStore, nil
	}

	StateStore.LoadState()
	return StateStore, nil
}

func (s *StuiState) SaveState() error {
	if !s.active {
		logger.Debugf("state store: path is null, state not active, no state saved")
		return nil
	}

	file, err := os.Create(s.statePath)
	if err != nil {
		return fmt.Errorf("failed to open state file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(s.State); err != nil {
		return fmt.Errorf("failed to encode state to JSON: %w", err)
	}
	return nil
}

func (s *StuiState) LoadState() error {
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		return fmt.Errorf("store: failed to read state file: %w", err)
	}
	var state StuiStateStruct
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state JSON: %w", err)
	}
	s.State = state
	return nil
}
