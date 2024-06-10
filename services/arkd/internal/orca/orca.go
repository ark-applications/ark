package orca

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dkimot/ark/services/arkd/internal/arkd"
	"github.com/dkimot/ark/services/arkd/internal/config"
	"github.com/dkimot/ark/services/arkd/internal/proxy"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

var ErrInsufficientResourcesAvailable = errors.New("orca: cannot schedule task, insufficient resources available")

type Orchestrator interface {
	ListTasks(ctx context.Context) ([]arkd.Task, error)
	InspectTask(ctx context.Context, taskId ulid.ULID) (*arkd.Task, error)

	StartTask(ctx context.Context, taskDef arkd.TaskDefinition) ([]byte, error)
	StopTask(ctx context.Context, taskId ulid.ULID, signal string) error
	WakeTask(ctx context.Context, taskId ulid.ULID) error
	DestroyTask(ctx context.Context, taskId ulid.ULID, force bool) error
}

func Start(cfg config.Config, logger zerolog.Logger, moby *docker.Client, taskStore *arkd.TaskStore, pxy proxy.Proxy) (Orchestrator, error) {
  o := &Orca{cfg: cfg, l: logger, moby: moby, taskStore: taskStore, proxy: pxy}

	go o.startWatcher()

	return o, nil
}

type Orca struct {
	cfg       config.Config
  l         zerolog.Logger
	moby      *docker.Client
	mtx       sync.Mutex
	taskStore *arkd.TaskStore
  proxy     proxy.Proxy
}

func (o *Orca) startWatcher() {
	for range time.Tick(time.Second) {
		ctx := context.Background()

		containers, err := o.moby.ContainerList(ctx, container.ListOptions{All: true})
		if err != nil {
			panic(err)
		}

		for _, ctr := range containers {
			if ctr.Labels["arkd"] == "1" && len(ctr.Names) > 0 {
				taskId := ulid.MustParse(ctr.Names[0][1:])
				task, err := o.taskStore.GetTask(ctx, taskId)
				if err != nil {
					panic(err)
				}

				// update status
				var taskStatus arkd.TaskStatus
				switch ctr.Status {
				case "Created":
					taskStatus = arkd.TaskStatusStarting
				}

				if strings.HasPrefix(ctr.Status, "Exited") {
					taskStatus = arkd.TaskStatusExited
				}

				if err := o.taskStore.SetTaskStatus(ctx, task, taskStatus); err != nil {
					panic(err)
				}
			}
		}
	}
}

func (o *Orca) DestroyTask(ctx context.Context, taskId ulid.ULID, force bool) error {
	task, err := o.taskStore.GetTask(ctx, taskId)
	if err != nil {
		return err
	}

  if task.ContainerID == "" {
    if err := o.taskStore.DeleteTask(ctx, taskId); err != nil {
      return err
    }

    return nil
  }

	if err := o.moby.ContainerStop(ctx, task.ContainerID, container.StopOptions{}); err != nil {
    return fmt.Errorf("could not stop container %s: %w", task.ContainerID, err)
	}

	if err := o.moby.ContainerRemove(ctx, task.ContainerID, container.RemoveOptions{}); err != nil {
    return fmt.Errorf("could not remove container %s: %w", task.ContainerID, err)
	}

  if err := o.proxy.DelistApp(task.ID.String()); err != nil {
    return err
  }

	if err := o.taskStore.DeleteTask(ctx, taskId); err != nil {
		return err
	}

	return nil
}

func (o *Orca) InspectTask(ctx context.Context, taskId ulid.ULID) (*arkd.Task, error) {
	panic("unimplemented")
}

func (o *Orca) ListTasks(ctx context.Context) ([]arkd.Task, error) {
	panic("unimplemented")
}

func (o *Orca) StartTask(ctx context.Context, taskDef arkd.TaskDefinition) ([]byte, error) {
  startedAt := time.Now()
  defer func()  {
    o.l.Debug().
      Str("app_name", taskDef.AppName).
      Str("deployment_name", taskDef.DeploymentName).
      Str("stack_name", taskDef.StackName).
      Dur("took", time.Since(startedAt)).
      Msg("starting task")
  }()

	// set taskdef defaults
	if taskDef.Cpu == 0.0 {
		taskDef.Cpu = o.cfg.DefaultTaskCpu
	}
	if taskDef.Memory == 0 {
		taskDef.Memory = o.cfg.DefaultTaskMem
	}

	o.mtx.Lock()
	defer o.mtx.Unlock()

	// verify capacity
	if arkd.GetSystemMetrics(ctx, o.taskStore).AvailableCpu < taskDef.Cpu {
		return nil, ErrInsufficientResourcesAvailable
	}

  return startTask(ctx, taskDef, o.moby, o.taskStore, o.proxy)
}

func (o *Orca) StopTask(ctx context.Context, taskId ulid.ULID, signal string) error {
	// verify task exists

	// stop container

	// update task status

	panic("unimplemented")
}

func (o *Orca) WakeTask(ctx context.Context, taskId ulid.ULID) error {
	// find stopped task

	// start container

	// update task status

	panic("unimplemented")
}
