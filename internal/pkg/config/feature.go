package config

type Feature struct {
	Metrics   bool `mapstructure:"metrics" json:"metrics" yaml:"metrics"`
	Profiling bool `mapstructure:"profiling" json:"profiling" yaml:"profiling"`
	ApiDoc    bool `mapstructure:"apiDoc" json:"apiDoc" yaml:"apiDoc"`
	QueueDash bool `mapstructure:"queueDash" json:"queueDash" yaml:"queueDash"`
}
