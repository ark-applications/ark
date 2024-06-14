package models

type Disk struct {
  Model
  WorkerID string
  Size int
  DeploymentID uint
}
