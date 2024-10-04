package config

type Redis struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"-" yaml:"password"`
	Database int    `mapstructure:"database" json:"database" yaml:"database"`
}
