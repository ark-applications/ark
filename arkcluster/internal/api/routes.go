package api

import (
	"encoding/json"
	"net/http"

	"github.com/dkimot/ark"
	"github.com/dkimot/ark/arkcluster/internal/config"
	"github.com/dkimot/ark/arkcluster/internal/dao"
	"github.com/dkimot/ark/arkcluster/internal/usecase"
	"github.com/dkimot/ark/arkcluster/internal/models"
	"gorm.io/gorm"
)

func addRoutes(
  mux *http.ServeMux,
  config config.Config,
  db *gorm.DB,
) {

  // stack routes
  mux.Handle("GET /v1/stacks", handleV1ListStacks(db))
  mux.Handle("POST /v1/stacks", handleV1CreateStack(db))
  mux.Handle("GET /v1/stacks/{stackName}", handleV1GetStack(db))
  mux.Handle("PUT /v1/stacks/{stackName}/definition", handleV1UpsertStackDefinition(db))
  mux.Handle("PUT /v1/stacks/{stackName}/secrets", notImplementedHandler())
  mux.Handle("DELETE /v1/stacks/{stackName}", handleV1DeleteStack(db))

  // deployment routes
  mux.Handle("GET /v1/stacks/{stackName}/deployments", handleV1ListDeployments(db))
  mux.Handle("POST /v1/stacks/{stackName}/deployments", handleV1CreateDeployment(db))
  mux.Handle("GET /v1/stacks/{stackName}/deployments/{deploymentName}", notImplementedHandler())
  mux.Handle("POST /v1/deployments/{deploymentId}/deploy", notImplementedHandler())
  mux.Handle("PUT /v1/deployments/{deploymentId}/stack_definition", notImplementedHandler())
  mux.Handle("PUT /v1/deployments/{deploymentId}/secrets", notImplementedHandler())
  mux.Handle("DELETE /v1/deployments/{deploymentId}", notImplementedHandler())

  // basic healthcheck
  mux.Handle("GET /v1/up", handleV1HealthCheck(config))
}

func notImplementedHandler() http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotImplemented)
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
		})
	})
}

func handleV1ListStacks(db *gorm.DB) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    var stacks []models.Stack
    result := db.Find(&stacks)
    if result.Error != nil {
      renderErr(w, r, result.Error)
      return
    }

    encode(w, r, http.StatusOK, stacks)
  })
}

func handleV1CreateStack(db *gorm.DB) http.Handler {
  type request struct {
    Name string `json:"name"`
  }

  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    var body request
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
      renderErr(w, r, err)
      return
    }

    stack := models.Stack{
      Name: body.Name,
    }

    result := db.Create(&stack)
    if result.Error != nil {
      renderErr(w, r, result.Error)
      return
    }

    encode(w, r, http.StatusCreated, stack)
  })
}

func handleV1GetStack(db *gorm.DB) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    stackName := r.PathValue("stackName")

    stack, err := dao.GetStackByName(r.Context(), db, stackName)
    if err != nil {
      renderErr(w, r, err)
      return
    }

    encode(w, r, http.StatusOK, stack)
  })
}

func handleV1DeleteStack(db *gorm.DB) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    stackName := r.PathValue("stackName")

    var stack models.Stack
    result := db.First(&stack, "name = ?", stackName)
    if result.Error != nil {
      renderErr(w, r, result.Error)
      return
    }

    result = db.Delete(&stack, stack.ID)
    if result.Error != nil {
      renderErr(w, r, result.Error)
      return
    }

    w.WriteHeader(http.StatusAccepted)
  })
}

func handleV1UpsertStackDefinition(db *gorm.DB) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    var body ark.StackDefinition
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
      renderErr(w, r, err)
      return
    }

    if err := usecase.UpsertStackDefinition(r.Context(), db, r.PathValue("stackName"), body); err != nil {
      renderErr(w, r, err)
      return
    }

    w.WriteHeader(http.StatusAccepted)
  })
}

func handleV1ListDeployments(db *gorm.DB) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    stack, err := getStackFromPath(r, db)
    if err != nil {
      renderErr(w, r, err)
      return
    }

    var deployments []models.Deployment
    result := db.Find(&deployments, "stack_id = ?", stack.ID)
    if result.Error != nil {
      renderErr(w, r, result.Error)
      return
    }

    encode(w, r, http.StatusOK, deployments)
  })
}

func handleV1CreateDeployment(db *gorm.DB) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    stack, err := getStackFromPath(r, db)
    if err != nil {
      renderErr(w, r, err)
      return
    }

    
  })
}

func getStackFromPath(r *http.Request, db *gorm.DB) (*models.Stack, error) {
  stackName := r.PathValue("stackName")
  return dao.GetStackByName(r.Context(), db, stackName)
}
