package config

import "time"

type Server struct {
	Name     string `mapstructure:"name" json:"name" yaml:"name"`
	Mode     string `mapstructure:"mode" json:"mode" yaml:"mode"`
	Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`
	Timezone string `mapstructure:"timezone" json:"timezone" yaml:"timezone"`
}

func (a Server) SetTimezone() {
	time.Local, _ = time.LoadLocation(a.Timezone)
}

func (a Server) IsProduction() bool {
	return a.Mode == "release"
}
