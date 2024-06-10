package arkd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
	"go.etcd.io/bbolt"
)

var tasksBucketName = []byte("TasksBucket")

var ErrTaskNotFound = errors.New("task not found")
var ErrNilTask = errors.New("nil task")

func NewTaskStore(db *bbolt.DB, logger zerolog.Logger) (*TaskStore, error) {
	// ensure tasks bucket exists
	err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket(tasksBucketName)
		if err != nil && err != bbolt.ErrBucketExists {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	taskStore := &TaskStore{
		db:         db,
		logger:     logger,
		aggMetrics: &AggTaskMetrics{},
	}

	tasks, err := taskStore.GetTasks(context.Background())
	if err != nil {
		return nil, err
	}
	if err := taskStore.updateAggMetrics(tasks); err != nil {
		return nil, err
	}

	return taskStore, nil
}

type TaskStore struct {
	db         *bbolt.DB
	logger     zerolog.Logger
	metricsMtx sync.RWMutex
	aggMetrics *AggTaskMetrics
}

// heads up, this locks the metrics mutex and is potentially called
// within a bolt tx. this might cause deadlocks
func (ts *TaskStore) updateAggMetrics(tasks []Task) error {
	allocCpu := 0.0
	allocMem := 0

	for _, t := range tasks {
		allocCpu += t.CPU
		allocMem += t.Memory
	}

	ts.metricsMtx.Lock()
	defer ts.metricsMtx.Unlock()

	ts.aggMetrics.TotalTasks = len(tasks)
	ts.aggMetrics.AllocatedCpu = allocCpu
	ts.aggMetrics.AllocatedMem = allocMem

	return nil
}

func (ts *TaskStore) CreateTask(ctx context.Context, taskDef TaskDefinition) (*Task, error) {
	t, err := NewTask(taskDef)
	if err != nil {
		return nil, err
	}

	err = ts.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(tasksBucketName)

		buf, err := writeTaskBytes(*t)
		if err != nil {
			return err
		}

		err = b.Put(t.ID.Bytes(), buf)
		if err != nil {
			return err
		}

		tasks, err := listTasksFromBucket(tx)
		if err != nil {
			return err
		}

		return ts.updateAggMetrics(tasks)
	})

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (ts *TaskStore) DeleteTask(ctx context.Context, id ulid.ULID) error {
	return ts.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(tasksBucketName)

		if err := b.Delete(id.Bytes()); err != nil {
			return err
		}

		tasks, err := listTasksFromBucket(tx)
		if err != nil {
			return err
		}

		return ts.updateAggMetrics(tasks)
	})
}

func (ts *TaskStore) GetTask(ctx context.Context, taskId ulid.ULID) (*Task, error) {
	var task *Task
	err := ts.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(tasksBucketName)

		buf := b.Get(taskId.Bytes())
		t, err := readTaskBytes(buf)
		if err != nil {
			return err
		}

		task = &t
		return nil
	})
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (ts *TaskStore) GetTasks(ctx context.Context) ([]Task, error) {
	tasks := make([]Task, 0)
	err := ts.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(tasksBucketName)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			task, err := readTaskBytes(v)
			if err != nil {
				return err
			}

			tasks = append(tasks, task)
		}

		return nil
	})

	return tasks, err
}

func (ts *TaskStore) SetTaskStatus(ctx context.Context, task *Task, status TaskStatus) error {
	return ts.db.Update(func(tx *bbolt.Tx) error {
		taskId := task.ID.Bytes()

		b := tx.Bucket(tasksBucketName)

		taskB := b.Get(taskId)

		taskB[0] = []byte(strconv.Itoa(int(status)))[0]

		task.Status = status
		return b.Put(taskId, taskB)
	})
}

func (ts *TaskStore) UpdateTask(ctx context.Context, task *Task) error {
	return ts.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(tasksBucketName)

		buf, err := writeTaskBytes(*task)
		if err != nil {
			return err
		}

		err = b.Put(task.ID.Bytes(), buf)
		if err != nil {
			return err
		}

		tasks, err := listTasksFromBucket(tx)
		if err != nil {
			return err
		}

		return ts.updateAggMetrics(tasks)
	})
}

func (ts *TaskStore) AggMetrics(ctx context.Context) *AggTaskMetrics {
	ts.metricsMtx.RLock()
	defer ts.metricsMtx.RUnlock()

	return ts.aggMetrics
}

func listTasksFromBucket(tx *bbolt.Tx) ([]Task, error) {
	tasks := make([]Task, 0)
	b := tx.Bucket(tasksBucketName)
	c := b.Cursor()

	for k, v := c.First(); k != nil; k, v = c.Next() {
		task, err := readTaskBytes(v)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func readTaskBytes(raw []byte) (Task, error) {
  if raw == nil {
    return Task{}, ErrNilTask
  }

	status, err := strconv.ParseInt(string(raw[0]), 10, 64)
	if err != nil {
		return Task{}, err
	}

	var task Task
	if err := json.Unmarshal(raw[1:], &task); err != nil {
		return Task{}, err
	}

	task.Status = TaskStatus(status)

	return task, nil
}

func writeTaskBytes(task Task) ([]byte, error) {
	status := task.Status
	taskStr, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("writeTaskBytes: %w", err)
	}

	s := strconv.Itoa(int(status))
	buf := append([]byte(s), taskStr...)

	return buf, nil
}
