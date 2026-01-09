package config

type ColorConfig struct {
	Primary       ConfigString `yaml:"primary,omitempty"`
	Yellow        ConfigString `yaml:"yellow,omitempty"`
	Green         ConfigString `yaml:"green,omitempty"`
	Gray          ConfigString `yaml:"gray,omitempty"`
	Blue          ConfigString `yaml:"blue,omitempty"`
	White         ConfigString `yaml:"white,omitempty"`
	Black         ConfigString `yaml:"black,omitempty"`
	Red           ConfigString `yaml:"red,omitempty"`
	Magenta       ConfigString `yaml:"magenta,omitempty"`
	Cyan          ConfigString `yaml:"cyan,omitempty"`
	BrightBlack   ConfigString `yaml:"bright-black,omitempty"`
	BrightRed     ConfigString `yaml:"bright-red,omitempty"`
	BrightGreen   ConfigString `yaml:"bright-green,omitempty"`
	BrightYellow  ConfigString `yaml:"bright-yellow,omitempty"`
	BrightBlue    ConfigString `yaml:"bright-blue,omitempty"`
	BrightMagenta ConfigString `yaml:"bright-magenta,omitempty"`
	BrightCyan    ConfigString `yaml:"bright-cyan,omitempty"`
	BrightWhite   ConfigString `yaml:"bright-white,omitempty"`
}

func emptyColorConfig() ColorConfig {
	return ColorConfig{
		Primary:       "",
		Yellow:        "",
		Green:         "",
		Gray:          "",
		Blue:          "",
		White:         "",
		Black:         "",
		Red:           "",
		Magenta:       "",
		Cyan:          "",
		BrightBlack:   "",
		BrightRed:     "",
		BrightGreen:   "",
		BrightYellow:  "",
		BrightBlue:    "",
		BrightMagenta: "",
		BrightCyan:    "",
		BrightWhite:   "",
	}
}
