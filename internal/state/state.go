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
	statePath string
	state     map[string]string
	active    bool
}

// Sets up state file and parent directories if they don't yet exist.
// Loads state from pre-existing files if available.
func InitializeStuiState(basePath string) (*StuiState, error) {
	StateStore := &StuiState{
		statePath: path.Join(basePath, "stui-state.json"),
		state:     make(map[string]string),
		active:    basePath != "",
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

func (s StuiState) GetStateKey(key string) (string, bool) {
	val, found := s.state[key]
	return val, found
}

func (s *StuiState) SetStateKey(k, v string) {
	s.state[k] = v
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

	if err := encoder.Encode(s.state); err != nil {
		return fmt.Errorf("failed to encode state to JSON: %w", err)
	}
	return nil
}

func (s *StuiState) LoadState() error {
	data, err := os.ReadFile(s.statePath)
	if err != nil {
		return fmt.Errorf("store: failed to read state file: %w", err)
	}
	var state map[string]string
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state JSON: %w", err)
	}
	s.state = state
	return nil
}
