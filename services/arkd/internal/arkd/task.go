package arkd

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type TaskStatus int

const (
	TaskStatusUnknown TaskStatus = iota
	TaskStatusPending
	TaskStatusImagePull
	TaskStatusCreating
	TaskStatusStarting
	TaskStatusRunning
	TaskStatusSuspended
	TaskStatusExited
	TaskStatusCrashed
)

type TaskDefinition struct {
	AppName        string  `json:"app_name"`
	DeploymentName string  `json:"deployment_name"`
	StackName      string  `json:"stack_name"`
	Image          string  `json:"image"`
	HealthCheck    string  `json:"health_check"`
	Cpu            float64 `json:"cpu"`
	Memory         int     `json:"memory"`
}

func NewTask(taskDef TaskDefinition) (*Task, error) {
	imageRef, err := NewImageRef(taskDef.Image)
	if err != nil {
		return nil, err
	}

	return &Task{
		ID:     ulid.Make(),
    AppName: taskDef.AppName,
    DeploymentName: taskDef.DeploymentName,
    StackName: taskDef.StackName,
		Status: TaskStatusPending,
		CPU:    taskDef.Cpu,
		Memory: taskDef.Memory,
		Image:  imageRef,
	}, nil
}

type Task struct {
	ID             ulid.ULID  `json:"id"`
	AppName        string     `json:"app_name"`
	DeploymentName string     `json:"deployment_name"`
	StackName      string     `json:"stack_name"`
	ContainerID    string     `json:"container_id"`
	CPU            float64    `json:"cpu"`
	StartedAt      time.Time  `json:"started_at"`
	Status         TaskStatus `json:"status"`
	Memory         int        `json:"memory"`
	Image          ImageRef   `json:"image"`
}

type AggTaskMetrics struct {
	TotalTasks   int     `json:"total_tasks"`
	AllocatedCpu float64 `json:"allocated_cpu"`
	AllocatedMem int     `json:"allocated_mem"`
}
