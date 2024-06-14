package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dkimot/ark"
	"github.com/dkimot/ark/arkcluster/internal/models"
	"github.com/dkimot/ark/arkd"
	"gorm.io/gorm"
)

func DeployDeployment(ctx context.Context, db *gorm.DB, arkd arkd.Client, deployment *models.Deployment) error {
  var stack models.Stack
  result := db.First(&stack, deployment.StackID)
  if result.Error != nil {
    return result.Error
  }
  stackName := stack.Name

  // get definition
  var definition ark.StackDefinition
  if err := json.Unmarshal(deployment.StackDefRaw, &definition); err != nil {
    return fmt.Errorf("invalid stack definition on deployment %d: %w", deployment.ID, err)
  }

  // build images

  // find worker that can run this deployment

  // request deployment on worker
  currTasks, err := arkd.ListTasks(ctx, deployment.Name)
  if err != nil {
    return err
  }
  firstDeploy := len(currTasks) == 0
  var errWhileDeploying error
  defer func() {
    if errWhileDeploying != nil {
      arkd.DeleteDeployment(ctx, deployment.Name)
    }
  }()

  if firstDeploy {
    if err := deployServices(ctx, definition.Services, arkd, deployment.Name, stackName); err != nil {
      errWhileDeploying = err
      return err
    }
  }

  if err := deployApps(ctx, firstDeploy, definition.Apps, arkd, deployment.Name, stackName); err != nil {
    errWhileDeploying = err
    return err
  }

  return nil
}

func deployApps(ctx context.Context, firstDeploy bool, apps map[string]ark.AppDefinition, arkd arkd.Client, deploymentName, stackName string) error {
  for appName, appDef := range apps {
    appDef.Name = appName
    if firstDeploy {
      if err := createApp(ctx, arkd, appDef, "image", deploymentName, stackName); err != nil {
        return err
      }
    } else {
      if err := redeployApp(ctx, arkd, appDef); err != nil {
        return err
      }
    }
  }

  return nil
}

func deployServices(ctx context.Context, services map[string]ark.ServiceDefinition, arkd arkd.Client, deploymentName, stackName string) error {
  for srvName, srvDef := range services {
    srvDef.Name = srvName
    if err := createService(ctx, arkd, srvDef, deploymentName, stackName); err != nil {
      return err
    }
  }

  return nil
}

func createApp(ctx context.Context, client arkd.Client, appDef ark.AppDefinition, image, deploymentName, stackName string) error {
  var expPorts []string
  if appDef.Type == "web" {
    expPorts = []string{appDef.HttpService.ContainerPort}
  }

  return client.CreateTask(ctx, arkd.CreateTaskParams{
    AppName: appDef.Name,
    DeploymentName: deploymentName,
    StackName: stackName,
    Image: image,
    ExposedPorts: expPorts,
  })
}


func redeployApp(ctx context.Context, arkd arkd.Client, appDef ark.AppDefinition) error {
  return nil
}

func createService(ctx context.Context, client arkd.Client, srvDef ark.ServiceDefinition, deploymentName, stackName string) error {
  return nil
}
