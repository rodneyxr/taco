package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	shuffleMembers = func(n int, swap func(i, j int)) {}
	os.Exit(m.Run())
}

func TestCommandGeneratesMessageAndPersistsRotation(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "taco.yaml")
	statePath := filepath.Join(dir, "state.yaml")

	config := `tacos: 2
teams:
  default:
    - alice
    - bob
    - charles
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	first := runCommand(t, "--config", configPath, "--state", statePath, "--message", "nice work")
	if strings.TrimSpace(first) != "@alice @bob nice work 🌮" {
		t.Fatalf("first output = %q", first)
	}

	second := runCommand(t, "--config", configPath, "--state", statePath, "--message", "nice work")
	if strings.TrimSpace(second) != "@charles @alice nice work 🌮" {
		t.Fatalf("second output = %q", second)
	}
}

func TestCommandSupportsNoRotate(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "taco.yaml")
	statePath := filepath.Join(dir, "state.yaml")

	config := `teams:
  default:
    - alice
    - bob
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	first := runCommand(t, "--config", configPath, "--state", statePath, "--number-of-tacos", "1", "--no-rotate")
	second := runCommand(t, "--config", configPath, "--state", statePath, "--number-of-tacos", "1", "--no-rotate")

	if strings.TrimSpace(first) != "@alice 🌮" {
		t.Fatalf("first output = %q", first)
	}
	if first != second {
		t.Fatalf("no-rotate outputs differed: %q != %q", first, second)
	}
}

func TestCommandSupportsAttachedShortTacoCount(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "taco.yaml")
	statePath := filepath.Join(dir, "state.yaml")

	config := `teams:
  default:
    - alice
    - bob
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	output := runCommand(t, "--config", configPath, "--state", statePath, "-n2", "--no-rotate")

	if strings.TrimSpace(output) != "@alice @bob 🌮" {
		t.Fatalf("output = %q", output)
	}
}

func TestCommandWarnsWhenTacosExceedUniqueMembers(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "taco.yaml")
	statePath := filepath.Join(dir, "state.yaml")

	config := `teams:
  default:
    - alice
    - "@alice"
    - bob
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdout, stderr := runCommandWithStreams(t, "--config", configPath, "--state", statePath, "--number-of-tacos", "5")

	if strings.TrimSpace(stdout) != "@alice @bob 🌮" {
		t.Fatalf("stdout = %q", stdout)
	}
	wantWarning := `warning: team "default" only has 2 unique member(s); 3 taco(s) were not given`
	if strings.TrimSpace(stderr) != wantWarning {
		t.Fatalf("stderr = %q, want %q", stderr, wantWarning)
	}
}

func runCommand(t *testing.T, args ...string) string {
	t.Helper()

	cmd := NewRootCommand(BuildInfo{Version: "test", Commit: "abc", Date: "today"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v\n%s", err, out.String())
	}

	return out.String()
}

func runCommandWithStreams(t *testing.T, args ...string) (string, string) {
	t.Helper()

	cmd := NewRootCommand(BuildInfo{Version: "test", Commit: "abc", Date: "today"})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	return stdout.String(), stderr.String()
}
