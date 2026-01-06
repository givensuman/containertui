package config

type ConfigString string

func (cs ConfigString) IsAssigned() bool {
	return cs != ""
}

type ConfigBool bool
