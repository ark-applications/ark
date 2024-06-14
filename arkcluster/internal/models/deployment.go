package models

type Deployment struct {
  Model
  Name string
  StackID uint
  StackDefRaw []byte
  DeployedFor string
}
