package taco

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// State stores rotation order and offsets per team.
type State struct {
	Teams map[string]TeamState `yaml:"teams"`
}

// TeamState stores the shuffled rotation order and next offset for one team.
type TeamState struct {
	Index int      `yaml:"index"`
	Order []string `yaml:"order,omitempty"`
}

// UnmarshalYAML supports both the current structured state and the previous
// state format where each team value was only an integer index.
func (s *TeamState) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		return value.Decode(&s.Index)
	}

	type teamState TeamState
	return value.Decode((*teamState)(s))
}

// NewState returns an empty rotation state.
func NewState() State {
	return State{Teams: map[string]TeamState{}}
}

// DefaultStatePath returns the default rotation state path.
func DefaultStatePath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "taco", "state.yaml")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".taco-state.yaml")
	}
	return ".taco-state.yaml"
}

// LoadState reads rotation state. Missing files are treated as empty state.
func LoadState(path string) (State, error) {
	state := NewState()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, nil
		}
		return state, fmt.Errorf("read state: %w", err)
	}
	if len(data) == 0 {
		return state, nil
	}
	if err := yaml.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("parse state: %w", err)
	}
	if state.Teams == nil {
		state.Teams = map[string]TeamState{}
	}
	return state, nil
}

// TeamState returns the rotation state for team.
func (s State) TeamState(team string) TeamState {
	return s.Teams[team]
}

// SetTeamState stores the rotation state for team.
func (s State) SetTeamState(team string, teamState TeamState) {
	if s.Teams == nil {
		s.Teams = map[string]TeamState{}
	}
	s.Teams[team] = teamState
}

// NextIndex returns the next rotation offset for team.
func (s State) NextIndex(team string) int {
	return s.Teams[team].Index
}

// SetNextIndex stores the next rotation offset for team.
func (s State) SetNextIndex(team string, index int) {
	if s.Teams == nil {
		s.Teams = map[string]TeamState{}
	}
	teamState := s.Teams[team]
	teamState.Index = index
	s.Teams[team] = teamState
}

// Save writes rotation state to disk.
func (s State) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}
