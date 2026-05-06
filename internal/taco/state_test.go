package taco

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadStateSupportsLegacyIndexState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")
	data := []byte(`teams:
  default: 2
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write state: %v", err)
	}

	state, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}

	want := TeamState{Index: 2}
	if !reflect.DeepEqual(state.TeamState("default"), want) {
		t.Fatalf("team state = %v, want %v", state.TeamState("default"), want)
	}
}

func TestSaveStateWritesTeamOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")
	state := NewState()
	state.SetTeamState("default", TeamState{
		Index: 1,
		Order: []string{"@alice", "@bob"},
	})

	if err := state.Save(path); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := LoadState(path)
	if err != nil {
		t.Fatalf("LoadState returned error: %v", err)
	}

	if !reflect.DeepEqual(loaded.TeamState("default"), state.TeamState("default")) {
		t.Fatalf("loaded team state = %v, want %v", loaded.TeamState("default"), state.TeamState("default"))
	}
}
