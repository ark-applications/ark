package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dkimot/ark/services/arkd/internal/arkd"
	"github.com/dkimot/ark/services/arkd/internal/config"
	"github.com/dkimot/ark/services/arkd/internal/orca"
	docker "github.com/docker/docker/client"
  "github.com/justinas/alice"
	"github.com/rs/zerolog"
  "github.com/rs/zerolog/hlog"
	"go.etcd.io/bbolt"
)

func StartHttpServer(
	logger zerolog.Logger,
	config config.Config,
	db *bbolt.DB,
	taskStore *arkd.TaskStore,
	moby *docker.Client,
	or orca.Orchestrator,
) error {
	mux := http.NewServeMux()

	addRoutes(mux, config, taskStore, or)

	var handler http.Handler = mux

	// add middleware
  handler = addLoggerMiddleware(logger, handler)

  addr := fmt.Sprintf(":%v", config.ApiPort)
	server := &http.Server{
		Handler:           handler,
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
	}

  logger.Info().Msgf("API server listening at %s", addr)
	return server.ListenAndServe()
}

func addLoggerMiddleware(logger zerolog.Logger, handler http.Handler) http.Handler {
  c := alice.New()

  c = c.Append(hlog.NewHandler(logger))
  c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
    hlog.FromRequest(r).Info().
      Str("method", r.Method).
      Stringer("url", r.URL).
      Int("status", status).
      Int("size", size).
      Dur("duration", duration).
      Msg("")
  }))
  c = c.Append(hlog.RemoteAddrHandler("ip"))
  c = c.Append(hlog.UserAgentHandler("user_agent"))
  c = c.Append(hlog.RefererHandler("referer"))
  c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))
  c = c.Append(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      hlog.FromRequest(r).Info().
        Str("method", r.Method).
        Stringer("url", r.URL).
        Msg("handling request")
      next.ServeHTTP(w, r)
    })
  })

  return c.Then(handler)
}
