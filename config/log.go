package config

// logConfig is a struct to store log configuration.
//
// Formatter is the log formatter to be used.
// Level is the log level to be used.
//
// Example:
//
//	formatter: "json"
//	level: "info"
type logConfig struct {
	Formatter string `mapstructure:"formatter"`
	Level     string `mapstructure:"level"`
}
