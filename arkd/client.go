package arkd

import "context"

type Client interface {
  GetCapacity(ctx context.Context) error
  ListTasks(ctx context.Context, deploymentName string) ([]Task, error)
  GetTask(ctx context.Context, taskId string) error
  CreateTask(context.Context, CreateTaskParams) error
  UpdateTask(ctx context.Context, taskId string, updateParams string) error
  DeleteTask(ctx context.Context, taskId string) error

  DeleteDeployment(ctx context.Context, deploymentName string) error
}

type client struct {}

func NewClient() Client {
  return nil
}

type CreateTaskParams struct {
  AppName string `json:"app_name"`
  DeploymentName string `json:"deployment_name"`
  StackName string `json:"stack_name"`
  Image string `json:"image"`
  ExposedPorts []string `json:"exposed_ports"`
}
