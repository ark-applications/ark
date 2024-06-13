package api

import (
	"net/http"

	"github.com/dkimot/ark/arkcluster/internal/config"
)

func addRoutes(
  mux *http.ServeMux,
  config config.Config,
) {
  // stack routes
  mux.Handle("GET /v1/stacks", notImplementedHandler())
  mux.Handle("POST /v1/stacks", notImplementedHandler())
  mux.Handle("GET /v1/stacks/{stackName}", notImplementedHandler())
  mux.Handle("POST /v1/stacks/{stackName}", notImplementedHandler())
  mux.Handle("PUT /v1/stacks/{stackName}/definition", notImplementedHandler())
  mux.Handle("PUT /v1/stacks/{stackName}/secrets", notImplementedHandler())
  mux.Handle("DELETE /v1/stacks/{stackName}", notImplementedHandler())

  // deployment routes
  mux.Handle("GET /v1/stacks/{stackName}/deployments", notImplementedHandler())
  mux.Handle("POST /v1/stacks/{stackName}/deployments", notImplementedHandler())
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
