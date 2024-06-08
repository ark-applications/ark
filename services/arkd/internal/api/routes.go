package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dkimot/ark/services/arkd/internal/arkd"
	"github.com/dkimot/ark/services/arkd/internal/config"
	"github.com/dkimot/ark/services/arkd/internal/orca"
	"github.com/oklog/ulid/v2"
	"github.com/rs/zerolog"
)

func addRoutes(
	mux *http.ServeMux,
	logger zerolog.Logger,
	config config.Config,
	taskStore *arkd.TaskStore,
	orc orca.Orchestrator,
) {
	// get the current capacity of the worker
	mux.Handle("GET /v1/capacity", handleV1CapacityGet(taskStore))
	// list tasks and their statuses
	mux.Handle("GET /v1/tasks", handleV1TaskList(taskStore))
	// create a new task
	mux.Handle("POST /v1/tasks", handleV1TaskCreate(orc))
	// get a specific task
	mux.Handle("GET /v1/tasks/{taskId}", handleV1TaskGet())
	// update a task definition
	mux.Handle("PUT /v1/tasks/{taskId}", handleV1TaskUpdate())
	// delete a task
	mux.Handle("DELETE /v1/tasks/{taskId}", handleV1TaskDelete(orc))
	// health check
	mux.Handle("GET /v1/up", handleV1HealthCheck(config))
	// get current volumes
	mux.Handle("GET /v1/volumes", handleV1VolumesGet())
}

func handleV1CapacityGet(taskStore *arkd.TaskStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sm := arkd.GetSystemMetrics(ctx, taskStore)

		encode(w, r, http.StatusOK, &sm)
	})
}

func handleV1HealthCheck(cfg config.Config) http.Handler {
	type response struct {
		ApiVersion string `json:"api_version"`
    WorkerId   string `json:"worker_id"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encode(w, r, http.StatusOK, &response{
			ApiVersion: cfg.ApiVersion,
      WorkerId: cfg.WorkerId,
		})
	})
}

func handleV1TaskCreate(orc orca.Orchestrator) http.Handler {
	type request struct {
    AppName string `json:"app_name"`
    StackName string `json:"stack_name"`
    DeploymentName string `json:"deployment_name"`
		Image string  `json:"image"`
		Cpu   float64 `json:"cpu"`
		Mem   int     `json:"mem"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var body request
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			renderErr(w, r, err)
			return
		}

		_, err := orc.StartTask(ctx, arkd.TaskDefinition{
			Image:  body.Image,
			Cpu:    body.Cpu,
			Memory: body.Mem,
      AppName: body.AppName,
      StackName: body.StackName,
      DeploymentName: body.DeploymentName,
		})
		if err != nil {
			renderErr(w, r, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
}

func handleV1TaskDelete(orc orca.Orchestrator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawTaskId := r.PathValue("taskId")
		taskId, err := ulid.Parse(rawTaskId)
		if err != nil {
			renderErr(w, r, fmt.Errorf("parsing task id: %w", err))
			return
		}

		orc.DestroyTask(r.Context(), taskId, false)

		w.WriteHeader(http.StatusOK)
	})
}

func handleV1TaskGet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
}

func handleV1TaskList(taskStore *arkd.TaskStore) http.Handler {
	type response struct {
		Tasks []arkd.Task `json:"tasks"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tasks, err := taskStore.GetTasks(ctx)
		if err != nil {
			renderErr(w, r, err)
			return
		}

    // TODO: filter by app, stack, and deployment names
    appNameFilter := r.URL.Query().Get("app_name")
    stackNameFilter := r.URL.Query().Get("stack_name")
    deploymentNameFilter := r.URL.Query().Get("deployment_name")
    if appNameFilter != "" || stackNameFilter != "" || deploymentNameFilter != "" {
      filteredTasks := make([]arkd.Task, 0)
      for _, t := range tasks {
        if appNameFilter != "" && t.AppName != appNameFilter {
          continue
        }

        if stackNameFilter != "" && t.StackName != stackNameFilter {
          continue
        }

        if deploymentNameFilter != "" && t.DeploymentName != deploymentNameFilter {
          continue
        }

        filteredTasks = append(filteredTasks, t)
      }

      tasks = filteredTasks
    }

		encode(w, r, http.StatusOK, &response{
			Tasks: tasks,
		})
	})
}

func handleV1TaskUpdate() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
}

func handleV1VolumesGet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
}
