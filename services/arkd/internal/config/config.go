package config

const (
	DefaultPort = 5500
	DefaultCpu  = 1.0             // 1 vCPU
	DefaultMem  = 256 / (1 << 20) // 256 MB
)

type Config struct {
	ApiVersion     string  `json:"api_version"`
	ApiPort        int     `json:"api_port"`
	DefaultTaskCpu float64 `json:"default_task_cpu"`
	DefaultTaskMem int     `json:"default_task_mem"`
  WorkerId       string  `json:"worker_id"`
}

type ConfigFn func(cfg *Config)

func NewConfig(options ...ConfigFn) Config {
	cfg := Config{
		ApiPort:        DefaultPort,
		DefaultTaskCpu: DefaultCpu,
		DefaultTaskMem: DefaultMem,
	}

	for _, opt := range options {
		opt(&cfg)
	}

	return cfg
}

func WithApiVersion(version string) ConfigFn {
	return func(cfg *Config) {
		cfg.ApiVersion = version
	}
}

func WithWorkerId(wid string) ConfigFn {
  return func(cfg *Config) {
    cfg.WorkerId = wid
  }
}
