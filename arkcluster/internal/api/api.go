package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dkimot/ark/arkcluster/internal/config"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func StartHttpServer(
  logger zerolog.Logger,
  config config.Config,
  db *gorm.DB,
) error {
  mux := http.NewServeMux()

  addRoutes(mux, config, db)

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
