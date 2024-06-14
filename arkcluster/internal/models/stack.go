package models

type Stack struct {
  Model
  Name string `json:"name"`
  CurrentVersionID uint `json:"current_version_id"`
  Definitions []StackDef `json:"definitions"`
}

type StackDef struct {
  Model
  StackID uint
  RootApp string
  RawDefinition []byte
}

type AppDefinition struct {
  Type string
  RepoUrl string
  Build AppBuildDefinition
  Env map[string]string
  HttpService AppHttpServiceDefinition
  HealthCheck AppHealthCheckDefinition
  Disks map[string]AppDiskDefinition
}

type AppBuildDefinition struct {
  Dockerfile string
  IgnoreFile string
  BuildTarget string
  Args map[string]string
}

type AppDeployDefinition struct {
  Command string
  ReleaseCommand string
}

type AppDiskDefinition struct {
  MountPath string
  Size string
}

type AppHealthCheckDefinition struct {
  GracePeriod string
  Interval string
  Timeout string
  Command string
  Request string
}

type AppHttpServiceDefinition struct {
  ContainerPort string
  KeepAlive bool
}

type ServiceDefinition struct {
  Image string
  Env map[string]string
}
