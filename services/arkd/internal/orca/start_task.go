package orca

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/dkimot/ark/services/arkd/internal/arkd"
	"github.com/dkimot/ark/services/arkd/internal/proxy"
	"github.com/dkimot/ark/services/arkd/pkg"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kozhurkin/pipers"
)

func startTask(
  ctx context.Context, 
  taskDef arkd.TaskDefinition, 
  moby *docker.Client, 
  taskStore *arkd.TaskStore,
  proxy     proxy.Proxy,
) ([]byte, error) {
	// add task to task storage
	task, err := taskStore.CreateTask(ctx, taskDef)
	if err != nil {
		return nil, err
	}

	if err := taskStore.SetTaskStatus(ctx, task, arkd.TaskStatusImagePull); err != nil {
		return nil, err
	}

	var networkId string
	desiredNetworkName := fmt.Sprintf("%s-%s-net", taskDef.DeploymentName, taskDef.StackName)

	pp := pipers.FromFuncs(
		// pull image
		func() (interface{}, error) {
			return nil, pullImage(ctx, task.Image.FullName, moby)
		},
		// ensure network is created
		func() (interface{}, error) {
      netId, err := findOrCreateNetwork(ctx, desiredNetworkName, moby)
      if err != nil {
        return nil, err
      }

      networkId = netId

			return nil, nil
		},
	)
	if _, err := pp.Resolve(); err != nil {
		return nil, err
	}

	if err := taskStore.SetTaskStatus(ctx, task, arkd.TaskStatusCreating); err != nil {
		return nil, err
	}

	// ----- create container -----

  // setupContainerPortMap also sets the HostPortBindings on the task.
  // this will get saved in the update task that occurs after container create
  portMap, err := setupContainerPortMap(ctx, task, taskDef.ExposedPorts, proxy)
  if err != nil {
    return nil, err
  }

	ccResp, err := moby.ContainerCreate(
    ctx, 
    &container.Config{
      AttachStdout: true,
      Image:        task.Image.FullName,
      Labels: map[string]string{
        "arkd": "1",
        "arkd_task_id": task.ID.String(),
        "arkd_qualified_name": task.QualifiedName(),
      },
    }, 
    &container.HostConfig{
      NetworkMode: "bridge",
      PortBindings: portMap,
    }, 
    nil,
    nil,
    task.ID.String(),
    )
	if err != nil {
    return nil, fmt.Errorf("could not create container: %w", err)
	}
	task.ContainerID = ccResp.ID
	task.Status = arkd.TaskStatusStarting
	if err := taskStore.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	if err := moby.NetworkConnect(ctx, networkId, ccResp.ID, nil); err != nil {
		return nil, fmt.Errorf("could not connect network: %w", err)
	}

	// start container
	if err := moby.ContainerStart(ctx, ccResp.ID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("could not start container: %w", err)
	}

	task.StartedAt = time.Now()
	task.Status = arkd.TaskStatusRunning
	if err := taskStore.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// return task id
	return task.ID.Bytes(), nil
}

func setupContainerPortMap(
  _ context.Context, 
  task *arkd.Task, 
  portsToExpose []string,
  proxy proxy.Proxy,
) (nat.PortMap, error) {
  portMap := make(nat.PortMap)
  if task.HostPortBindings == nil {
    task.HostPortBindings = make(map[string]string)
  }

  for _, containerPort := range portsToExpose {
    freePort, err := pkg.GetFreePort()
    if err != nil {
      return nil, fmt.Errorf("could not get free port: %w", err)
    }
    
    hostPort := strconv.Itoa(freePort)

    portMap[nat.Port(containerPort + "/tcp")] = []nat.PortBinding{{HostPort: hostPort}}
    task.HostPortBindings[hostPort] = containerPort
    proxy.RegisterApp(task.ID.String(), task.QualifiedName(), task.Domain(), hostPort)
  }

  return portMap, nil
}

func pullImage(ctx context.Context, imageName string, moby *docker.Client) error {
  rc, err := moby.ImagePull(ctx, imageName, image.PullOptions{})
  if err != nil {
    return err
  }
  defer rc.Close()

  if _, err := io.ReadAll(rc); err != nil {
    return err
  }

  return nil
}

func findOrCreateNetwork(ctx context.Context, desiredNetworkName string, moby *docker.Client) (string, error) {
			nets, err := moby.NetworkList(ctx, types.NetworkListOptions{
				Filters: filters.NewArgs(filters.Arg("name", desiredNetworkName)),
			})
			if err != nil {
				return "", err
			}

			if len(nets) > 0 {
        return nets[0].ID, nil
			}

			netCreateOpts := types.NetworkCreate{
				Driver:     "bridge",
				EnableIPv6: false,
				Labels:     map[string]string{"arkd": "1"},
			}
			netCreateResp, err := moby.NetworkCreate(ctx, desiredNetworkName, netCreateOpts)
			if err != nil {
        return "", fmt.Errorf("could not create network: %w", err)
			}

			return netCreateResp.ID, nil
}
