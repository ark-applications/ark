package config

const (
  DefaultPort = 4400
)

type Config struct {
  ApiVersion string `json:"api_version"`
  ApiPort    int    `json:"api_port"`
}

type ConfigOptFn func(cfg *Config)

func NewConfig(options ...ConfigOptFn) Config {
  cfg := Config{
    ApiPort: DefaultPort,
  }

  for _, opt := range options {
    opt(&cfg)
  }

  return cfg
}

func WithApiVersion(version string) ConfigOptFn {
	return func(cfg *Config) {
		cfg.ApiVersion = version
	}
}
