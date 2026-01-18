package config

type ThemeConfig struct {
	Primary  ConfigString `yaml:"primary,omitempty"`
	Border   ConfigString `yaml:"border,omitempty"`
	Text     ConfigString `yaml:"text,omitempty"`
	Muted    ConfigString `yaml:"muted,omitempty"`
	Selected ConfigString `yaml:"selected,omitempty"`
	Success  ConfigString `yaml:"success,omitempty"`
	Warning  ConfigString `yaml:"warning,omitempty"`
	Error    ConfigString `yaml:"error,omitempty"`
}

func emptyThemeConfig() ThemeConfig {
	return ThemeConfig{
		Primary:  "",
		Border:   "",
		Text:     "",
		Muted:    "",
		Selected: "",
		Success:  "",
		Warning:  "",
		Error:    "",
	}
}
