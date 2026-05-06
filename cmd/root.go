package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/rodneyxr/taco/internal/taco"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultTemplate = `{{ team "default" }} {{ .Message }} {{ .Emoji }}`

var shuffleMembers taco.ShuffleFunc = rand.Shuffle

// BuildInfo is populated by GoReleaser ldflags for published builds.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type options struct {
	configPath string
	count      int
	message    string
	emoji      string
	template   string
	team       string
	statePath  string
	noRotate   bool
}

// Execute runs the root command.
func Execute(info BuildInfo) {
	if err := NewRootCommand(info).Execute(); err != nil {
		os.Exit(1)
	}
}

// NewRootCommand builds the CLI root command.
func NewRootCommand(info BuildInfo) *cobra.Command {
	opts := options{}

	cmd := &cobra.Command{
		Use:     "taco",
		Short:   "Create HeyTaco Slack messages",
		Version: formatVersion(info),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, opts)
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")

	flags := cmd.Flags()
	flags.StringVarP(&opts.configPath, "config", "c", "taco.yaml", "config file (searches current directory, then home)")
	flags.IntVarP(&opts.count, "number-of-tacos", "n", 5, "number of teammates to include")
	flags.StringVarP(&opts.message, "message", "m", "", "message to include")
	flags.StringVar(&opts.emoji, "emoji", "🌮", "emoji to include")
	flags.StringVar(&opts.template, "template", defaultTemplate, "message template")
	flags.StringVarP(&opts.team, "team", "t", "default", "team to use with the default template")
	flags.StringVar(&opts.statePath, "state", "", "rotation state file")
	flags.BoolVar(&opts.noRotate, "no-rotate", false, "preview without advancing rotation")

	return cmd
}

func run(cmd *cobra.Command, opts options) error {
	cfg, err := loadConfig(opts.configPath, cmd.Flags().Changed("config"))
	if err != nil {
		return err
	}

	if !cmd.Flags().Changed("number-of-tacos") {
		opts.count = cfg.TacoCount(opts.count)
	}
	if !cmd.Flags().Changed("message") && cfg.Message != "" {
		opts.message = cfg.Message
	}
	if !cmd.Flags().Changed("emoji") && cfg.Emoji != "" {
		opts.emoji = cfg.Emoji
	}
	if !cmd.Flags().Changed("template") && cfg.Template != "" {
		opts.template = cfg.Template
	}
	if !cmd.Flags().Changed("template") && cmd.Flags().Changed("team") {
		opts.template = fmt.Sprintf(`{{ team %q }} {{ .Message }} {{ .Emoji }}`, opts.team)
	}

	if opts.count < 1 {
		return errors.New("number-of-tacos must be greater than zero")
	}
	if len(cfg.Teams) == 0 {
		return errors.New("no teams found; add a teams section to taco.yaml")
	}

	state := taco.NewState()
	statePath := opts.statePath
	if statePath == "" {
		statePath = taco.DefaultStatePath()
	}

	if !opts.noRotate {
		loadedState, err := taco.LoadState(statePath)
		if err != nil {
			return err
		}
		state = loadedState
	}

	usedTeams := map[string]bool{}
	output, err := taco.RenderWithTeam(opts.template, opts.message, opts.emoji, func(teamName string) (string, error) {
		if usedTeams[teamName] {
			return "", fmt.Errorf("team %q is used more than once in the template", teamName)
		}
		usedTeams[teamName] = true

		members, ok := cfg.Teams[teamName]
		if !ok {
			return "", fmt.Errorf("team %q was not found in config", teamName)
		}

		selected, next, err := taco.SelectRandomMembers(members, opts.count, state.TeamState(teamName), shuffleMembers)
		if err != nil {
			return "", fmt.Errorf("team %q: %w", teamName, err)
		}
		if missing := opts.count - len(selected); missing > 0 {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: team %q only has %d unique member(s); %d taco(s) were not given\n", teamName, len(selected), missing)
		}
		if !opts.noRotate {
			state.SetTeamState(teamName, next)
		}
		return strings.Join(selected, " "), nil
	})
	if err != nil {
		return err
	}
	if len(usedTeams) == 0 {
		return errors.New(`template must call {{ team "name" }} at least once`)
	}

	fmt.Fprintln(cmd.OutOrStdout(), output)

	if !opts.noRotate {
		return state.Save(statePath)
	}
	return nil
}

func formatVersion(info BuildInfo) string {
	if info.Version == "" {
		info.Version = "dev"
	}
	if info.Commit == "" {
		info.Commit = "none"
	}
	if info.Date == "" {
		info.Date = "unknown"
	}
	return fmt.Sprintf("taco %s (commit %s, built %s)", info.Version, info.Commit, info.Date)
}

func loadConfig(path string, explicit bool) (taco.Config, error) {
	cfg := taco.Config{}

	configPath, err := resolveConfigPath(path, explicit)
	if err != nil {
		return cfg, err
	}
	if configPath == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	v := viper.New()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func resolveConfigPath(path string, explicit bool) (string, error) {
	if path == "" {
		path = "taco.yaml"
	}

	if filepath.IsAbs(path) || strings.Contains(path, string(os.PathSeparator)) {
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) && !explicit {
				return "", nil
			}
			return "", fmt.Errorf("read config: %w", err)
		}
		return path, nil
	}

	candidates := []string{path}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, "."+path), filepath.Join(home, path))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("read config: %w", err)
		}
	}

	if explicit {
		return "", fmt.Errorf("read config: %s was not found", path)
	}
	return "", nil
}
