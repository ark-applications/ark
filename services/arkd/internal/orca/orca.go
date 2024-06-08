package orca

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/dkimot/ark/services/arkd/internal/arkd"
	"github.com/dkimot/ark/services/arkd/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	docker "github.com/docker/docker/client"
	"github.com/oklog/ulid/v2"
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

// Start will start an orchestrator that exposes
// a Nomad-like API to run tasks.
func Start(cfg config.Config, moby *docker.Client, taskStore *arkd.TaskStore) (Orchestrator, error) {
	o := &Orca{cfg: cfg, moby: moby, taskStore: taskStore}

	go o.startWatcher()

	return o, nil
}

type Orca struct {
	cfg       config.Config
	moby      *docker.Client
	mtx       sync.Mutex
	taskStore *arkd.TaskStore
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

	if err := o.moby.ContainerStop(ctx, task.ContainerID, container.StopOptions{}); err != nil {
		return err
	}

	if err := o.moby.ContainerRemove(ctx, task.ContainerID, container.RemoveOptions{}); err != nil {
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

	// add task to task storage
	task, err := o.taskStore.CreateTask(ctx, taskDef)
	if err != nil {
		return nil, err
	}

	if err := o.taskStore.SetTaskStatus(ctx, task, arkd.TaskStatusImagePull); err != nil {
		return nil, err
	}

	// pull image
	rc, err := o.moby.ImagePull(ctx, task.Image.FullName, image.PullOptions{})
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	if _, err := io.ReadAll(rc); err != nil {
		return nil, err
	}

	if err := o.taskStore.SetTaskStatus(ctx, task, arkd.TaskStatusCreating); err != nil {
		return nil, err
	}

	// create container
	ccResp, err := o.moby.ContainerCreate(ctx, &container.Config{
		AttachStdout: true,
		Image:        task.Image.FullName,
		Labels: map[string]string{
			"arkd": "1",
		},
	}, nil, nil, nil, task.ID.String())
	if err != nil {
		return nil, err
	}
	task.ContainerID = ccResp.ID
	task.Status = arkd.TaskStatusStarting
	if err := o.taskStore.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// start container
	if err := o.moby.ContainerStart(ctx, ccResp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	task.StartedAt = time.Now()
	task.Status = arkd.TaskStatusRunning
	if err := o.taskStore.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	// return task id
	return task.ID.Bytes(), nil
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
