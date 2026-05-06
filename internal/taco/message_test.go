package taco

import (
	"reflect"
	"testing"
)

func TestSelectMembersNormalizesAndRotates(t *testing.T) {
	selected, next, err := SelectMembers([]string{"@alice", "@bob", "charles"}, 2, 1)
	if err != nil {
		t.Fatalf("SelectMembers returned error: %v", err)
	}

	want := []string{"@bob", "@charles"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %v, want %v", selected, want)
	}
	if next != 0 {
		t.Fatalf("next = %d, want 0", next)
	}
}

func TestSelectMembersReturnsUniqueMembersWhenCountExceedsTeam(t *testing.T) {
	selected, next, err := SelectMembers([]string{"alice", "bob"}, 5, 0)
	if err != nil {
		t.Fatalf("SelectMembers returned error: %v", err)
	}

	want := []string{"@alice", "@bob"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %v, want %v", selected, want)
	}
	if next != 0 {
		t.Fatalf("next = %d, want 0", next)
	}
}

func TestSelectMembersSkipsDuplicateMentions(t *testing.T) {
	selected, next, err := SelectMembers([]string{"alice", "@alice", "bob", "charles"}, 2, 0)
	if err != nil {
		t.Fatalf("SelectMembers returned error: %v", err)
	}

	want := []string{"@alice", "@bob"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %v, want %v", selected, want)
	}
	if next != 3 {
		t.Fatalf("next = %d, want 3", next)
	}
}

func TestSelectRandomMembersUsesStoredShuffledRotation(t *testing.T) {
	selected, next, err := SelectRandomMembers(
		[]string{"alice", "bob", "charles"},
		2,
		TeamState{Order: []string{"@bob", "@charles", "@alice"}},
		identityShuffle,
	)
	if err != nil {
		t.Fatalf("SelectRandomMembers returned error: %v", err)
	}

	want := []string{"@bob", "@charles"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %v, want %v", selected, want)
	}
	if next.Index != 2 {
		t.Fatalf("next index = %d, want 2", next.Index)
	}
	if !reflect.DeepEqual(next.Order, []string{"@bob", "@charles", "@alice"}) {
		t.Fatalf("next order = %v", next.Order)
	}
}

func TestSelectRandomMembersReshufflesAtCycleBoundaryWithoutMessageDuplicates(t *testing.T) {
	selected, next, err := SelectRandomMembers(
		[]string{"alice", "bob", "charles"},
		2,
		TeamState{Index: 2, Order: []string{"@alice", "@bob", "@charles"}},
		identityShuffle,
	)
	if err != nil {
		t.Fatalf("SelectRandomMembers returned error: %v", err)
	}

	want := []string{"@charles", "@alice"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %v, want %v", selected, want)
	}

	wantOrder := []string{"@alice", "@bob", "@charles"}
	if !reflect.DeepEqual(next.Order, wantOrder) {
		t.Fatalf("next order = %v, want %v", next.Order, wantOrder)
	}
	if next.Index != 1 {
		t.Fatalf("next index = %d, want 1", next.Index)
	}
}

func TestSelectRandomMembersDefersAlreadySelectedMembersInNewCycle(t *testing.T) {
	selected, next, err := SelectRandomMembers(
		[]string{"alice", "bob", "charles"},
		2,
		TeamState{Index: 2, Order: []string{"@alice", "@bob", "@charles"}},
		reverseShuffle,
	)
	if err != nil {
		t.Fatalf("SelectRandomMembers returned error: %v", err)
	}

	want := []string{"@charles", "@bob"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %v, want %v", selected, want)
	}

	wantOrder := []string{"@bob", "@alice", "@charles"}
	if !reflect.DeepEqual(next.Order, wantOrder) {
		t.Fatalf("next order = %v, want %v", next.Order, wantOrder)
	}
	if next.Index != 1 {
		t.Fatalf("next index = %d, want 1", next.Index)
	}
}

func TestSelectRandomMembersReturnsUniqueMembersWhenCountExceedsTeam(t *testing.T) {
	selected, next, err := SelectRandomMembers(
		[]string{"alice", "@alice", "bob"},
		5,
		TeamState{},
		identityShuffle,
	)
	if err != nil {
		t.Fatalf("SelectRandomMembers returned error: %v", err)
	}

	want := []string{"@alice", "@bob"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %v, want %v", selected, want)
	}
	if next.Index != 2 {
		t.Fatalf("next index = %d, want 2", next.Index)
	}
}

func TestSelectRandomMembersReshufflesWhenSavedCycleIsExhausted(t *testing.T) {
	selected, next, err := SelectRandomMembers(
		[]string{"alice", "bob", "charles"},
		1,
		TeamState{Index: 3, Order: []string{"@alice", "@bob", "@charles"}},
		reverseShuffle,
	)
	if err != nil {
		t.Fatalf("SelectRandomMembers returned error: %v", err)
	}

	want := []string{"@charles"}
	if !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %v, want %v", selected, want)
	}

	wantOrder := []string{"@charles", "@bob", "@alice"}
	if !reflect.DeepEqual(next.Order, wantOrder) {
		t.Fatalf("next order = %v, want %v", next.Order, wantOrder)
	}
	if next.Index != 1 {
		t.Fatalf("next index = %d, want 1", next.Index)
	}
}

func TestRenderDefaultTemplate(t *testing.T) {
	got, err := Render(
		`{{ team "default" }} {{ .Message }} {{ .Emoji }}`,
		map[string][]string{"default": {"@alice", "@bob"}},
		"thanks for helping",
		"🌮",
	)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := "@alice @bob thanks for helping 🌮"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderOmitsEmptyMessageWhitespace(t *testing.T) {
	got, err := Render(
		`{{ team "default" }} {{ .Message }} {{ .Emoji }}`,
		map[string][]string{"default": {"@alice"}},
		"",
		"🌮",
	)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := "@alice 🌮"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderDefaultsBlankTeamName(t *testing.T) {
	got, err := Render(
		`{{ team "" }} {{ .Emoji }}`,
		map[string][]string{"default": {"@alice"}},
		"",
		"🌮",
	)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	want := "@alice 🌮"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func identityShuffle(n int, swap func(i, j int)) {}

func reverseShuffle(n int, swap func(i, j int)) {
	for i := 0; i < n/2; i++ {
		swap(i, n-1-i)
	}
}
