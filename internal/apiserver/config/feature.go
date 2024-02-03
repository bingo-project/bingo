package config

type Feature struct {
	ApiDoc    bool `mapstructure:"apiDoc" json:"apiDoc" yaml:"apiDoc"`
	QueueDash bool `mapstructure:"queueDash" json:"queueDash" yaml:"queueDash"`
}
