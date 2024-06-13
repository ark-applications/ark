package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dkimot/ark/arkcluster/internal/config"
	"github.com/rs/zerolog"
)

func StartHttpServer(
  logger zerolog.Logger,
  config config.Config,
) error {
  mux := http.NewServeMux()

  var handler http.Handler = mux

  // add middleware

  addr := fmt.Sprintf(":%v", config.ApiPort)
  server := &http.Server{
    Handler:           handler,
    Addr:              addr,
    ReadHeaderTimeout: 3 * time.Second,
  }

  logger.Info().Msgf("API server listening at %s", addr)
  return server.ListenAndServe()
}
