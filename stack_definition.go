package ark

type StackDefinition struct {
  FileVersion string `toml:"file_version"`
  Version string `toml:"version"`
  StackName string `toml:"stack"`
  RootApp string `toml:"root_app"`
  Apps   map[string]AppDefinition `toml:"apps"`
  Services map[string]ServiceDefinition `toml:"services"`
}

type AppDefinition struct {
  Name string `toml:"-"`
  Type string `toml:"type"`
  RepoUrl string `toml:"repo_url"`
  Build AppBuildDefinition `toml:"build"`
  Deploy AppDeployDefinition `toml:"deploy"`
  Env map[string]string `toml:"env"`
  HttpService AppHttpServiceDefinition `toml:"http_service"`
  HealthCheck AppHealthCheckDefinition `toml:"health_check"`
  Disks map[string]AppDiskDefinition `toml:"disks"`
}

type AppDeployDefinition struct {
  Command string `toml:"command"`
  ReleaseCommand string `toml:"release_command"`
}

type AppBuildDefinition struct {
  Dockerfile string `toml:"dockerfile"`
  Ignorefile string `toml:"ignorefile"`
  BuildTarget string `toml:"build_target"`
  Args map[string]string `toml:"args"`
}

type AppDiskDefinition struct {
  MountPath string `toml:"mount_path"`
  Size string `toml:"size"`
}

type AppHealthCheckDefinition struct {
  GracePeriod string `toml:"grace_period"`
  Interval string `toml:"interval"`
  Timeout string `toml:"timeout"`
  Command string `toml:"command"`
  Request string `toml:"request"`
}

type AppHttpServiceDefinition struct {
  ContainerPort string `toml:"container_port"`
  KeepAlive bool `toml:"keep_alive"`
}

type ServiceDefinition struct {
  Name string `toml:"-"`
  Image string `toml:"image"`
  RepoUrl string `toml:"repo_url"`
  Dockerfile string `toml:"dockerfile"`
  Env map[string]string `toml:"env"`
}
