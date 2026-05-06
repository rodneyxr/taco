package taco

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// ShuffleFunc shuffles n items by swapping indexes.
type ShuffleFunc func(n int, swap func(i, j int))

// Render turns a template and selected teams into a Slack-ready message.
func Render(template string, selections map[string][]string, message string, emoji string) (string, error) {
	return RenderWithTeam(template, message, emoji, func(name string) (string, error) {
		members, ok := selections[teamName(name)]
		if !ok {
			return "", fmt.Errorf("team %q was not found", name)
		}
		return strings.Join(members, " "), nil
	})
}

// RenderWithTeam executes a Go template with message, emoji, and team helpers.
func RenderWithTeam(templateText string, message string, emoji string, team func(name string) (string, error)) (string, error) {
	tmpl, err := template.New("message").Funcs(template.FuncMap{
		"team": func(name string) (string, error) {
			return team(teamName(name))
		},
	}).Parse(templateText)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := struct {
		Message string
		Emoji   string
	}{
		Message: strings.TrimSpace(message),
		Emoji:   strings.TrimSpace(emoji),
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}

	return strings.Join(strings.Fields(output.String()), " "), nil
}

// SelectMembers returns up to count unique members starting at start, plus the next index.
func SelectMembers(members []string, count int, start int) ([]string, int, error) {
	if len(members) == 0 {
		return nil, 0, fmt.Errorf("team has no members")
	}
	if count < 1 {
		return nil, 0, fmt.Errorf("count must be greater than zero")
	}

	start = normalizeIndex(start, len(members))
	selected := make([]string, 0, count)
	seen := make(map[string]bool, len(members))
	steps := 0
	for ; steps < len(members) && len(selected) < count; steps++ {
		member := NormalizeMention(members[(start+steps)%len(members)])
		if member == "" || seen[member] {
			continue
		}
		seen[member] = true
		selected = append(selected, member)
	}

	return selected, (start + steps) % len(members), nil
}

// SelectRandomMembers returns up to count unique members from a shuffled rotation.
func SelectRandomMembers(members []string, count int, state TeamState, shuffle ShuffleFunc) ([]string, TeamState, error) {
	if count < 1 {
		return nil, TeamState{}, fmt.Errorf("count must be greater than zero")
	}
	if shuffle == nil {
		return nil, TeamState{}, fmt.Errorf("shuffle function is required")
	}

	roster := UniqueMentions(members)
	if len(roster) == 0 {
		return nil, TeamState{}, fmt.Errorf("team has no members")
	}

	order, index := normalizeRotation(roster, state, shuffle)
	limit := count
	if limit > len(roster) {
		limit = len(roster)
	}

	selected := make([]string, 0, limit)
	selectedSet := make(map[string]bool, limit)
	for len(selected) < limit {
		if index >= len(order) {
			order = shuffledOrder(roster, selectedSet, shuffle)
			index = 0
		}

		member := order[index]
		index++
		if selectedSet[member] {
			continue
		}
		selectedSet[member] = true
		selected = append(selected, member)
	}

	return selected, TeamState{Index: index, Order: order}, nil
}

// UniqueMentions normalizes members and removes duplicate mentions while preserving order.
func UniqueMentions(members []string) []string {
	mentions := make([]string, 0, len(members))
	seen := make(map[string]bool, len(members))
	for _, member := range members {
		mention := NormalizeMention(member)
		if mention == "" || seen[mention] {
			continue
		}
		seen[mention] = true
		mentions = append(mentions, mention)
	}
	return mentions
}

// NormalizeMention strips any provided @ prefix and adds one Slack mention prefix.
func NormalizeMention(member string) string {
	member = strings.TrimLeft(strings.TrimSpace(member), "@")
	if member == "" {
		return ""
	}
	return "@" + member
}

func teamName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "default"
	}
	return name
}

func normalizeRotation(roster []string, state TeamState, shuffle ShuffleFunc) ([]string, int) {
	active := make(map[string]bool, len(roster))
	for _, member := range roster {
		active[member] = true
	}

	order := make([]string, 0, len(roster))
	seen := make(map[string]bool, len(roster))
	for _, member := range state.Order {
		member = NormalizeMention(member)
		if member == "" || !active[member] || seen[member] {
			continue
		}
		seen[member] = true
		order = append(order, member)
	}

	missing := make([]string, 0, len(roster)-len(order))
	for _, member := range roster {
		if seen[member] {
			continue
		}
		missing = append(missing, member)
	}
	shuffleStrings(missing, shuffle)
	order = append(order, missing...)

	if len(order) == 0 {
		order = shuffledOrder(roster, nil, shuffle)
	}

	return order, normalizeRotationIndex(state.Index, len(order))
}

func shuffledOrder(roster []string, deferred map[string]bool, shuffle ShuffleFunc) []string {
	first := make([]string, 0, len(roster))
	last := make([]string, 0, len(roster))
	for _, member := range roster {
		if deferred[member] {
			last = append(last, member)
			continue
		}
		first = append(first, member)
	}

	shuffleStrings(first, shuffle)
	shuffleStrings(last, shuffle)
	return append(first, last...)
}

func shuffleStrings(values []string, shuffle ShuffleFunc) {
	shuffle(len(values), func(i, j int) {
		values[i], values[j] = values[j], values[i]
	})
}

func normalizeRotationIndex(index int, length int) int {
	if index == length {
		return index
	}
	return normalizeIndex(index, length)
}

func normalizeIndex(index int, length int) int {
	index %= length
	if index < 0 {
		index += length
	}
	return index
}
