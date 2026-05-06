package taco

// Config mirrors taco.yaml.
type Config struct {
	Tacos         int                 `mapstructure:"tacos"`
	NumberOfTacos int                 `mapstructure:"numberOfTacos"`
	Message       string              `mapstructure:"message"`
	Emoji         string              `mapstructure:"emoji"`
	Template      string              `mapstructure:"template"`
	Teams         map[string][]string `mapstructure:"teams"`
}

// TacoCount returns the configured taco count, falling back to fallback.
func (c Config) TacoCount(fallback int) int {
	if c.NumberOfTacos > 0 {
		return c.NumberOfTacos
	}
	if c.Tacos > 0 {
		return c.Tacos
	}
	return fallback
}
