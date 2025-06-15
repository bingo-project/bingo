package config

type Websocket struct {
	Addr string `mapstructure:"addr" json:"addr" yaml:"addr"`
}
