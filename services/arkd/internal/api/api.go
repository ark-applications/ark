package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dkimot/ark/services/arkd/internal/arkd"
	"github.com/dkimot/ark/services/arkd/internal/config"
	"github.com/dkimot/ark/services/arkd/internal/orca"
	docker "github.com/docker/docker/client"
	"github.com/rs/zerolog"
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
	log.SetFlags(0)

	mux := http.NewServeMux()

	addRoutes(mux, logger, config, taskStore, or)

	var handler http.Handler = mux

	// add middleware

	server := &http.Server{
		Handler:           handler,
		Addr:              fmt.Sprintf(":%v", config.ApiPort),
		ReadHeaderTimeout: 3 * time.Second,
	}

	return server.ListenAndServe()
}
