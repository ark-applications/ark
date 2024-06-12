package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	docker "github.com/docker/docker/client"
	"go.etcd.io/bbolt"

	"github.com/dkimot/ark/services/arkd/internal/api"
	"github.com/dkimot/ark/services/arkd/internal/arkd"
	"github.com/dkimot/ark/services/arkd/internal/config"
	"github.com/dkimot/ark/services/arkd/internal/orca"
	"github.com/dkimot/ark/services/arkd/internal/proxy"
	"github.com/rs/zerolog"
)

var (
	ApiVersion = "alpha"
)

func run(ctx context.Context, _ []string, getenv func(string) string, _ io.Reader, stdout, _ io.Writer) (err error) {
	l := zerolog.New(stdout).With().
    Timestamp().
    Str("role", "arkd-worker").
    Logger()

  // Set up OpenTelemetry.
	otelShutdown, err := setupOtelSdk(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	wid, err := getWorkerId(getenv("ARKD_WORKER_ID"))
	if err != nil {
		return err
	}

	cfg := config.NewConfig(config.WithApiVersion(ApiVersion), config.WithWorkerId(wid))

	db, err := bbolt.Open("arkd.db", 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()

  pxy := proxy.New(cfg)

	taskStore, err := arkd.NewTaskStore(db, l)
	if err != nil {
		return err
	}

	moby, err := getDockerClient(ctx)
	if err != nil {
		return err
	}
	defer moby.Close()

	or, err := orca.Start(cfg, l, moby, taskStore, pxy)
	if err != nil {
		return err
	}

	return api.StartHttpServer(
		l,
		cfg,
		db,
		taskStore,
		moby,
		or,
	)
}

func getWorkerId(workerIdEnvVar string) (string, error) {
	hn, err := os.Hostname()
	if err != nil {
		return "", err
	}

	if workerIdEnvVar != "" {
		return hn + "-" + workerIdEnvVar, nil
	}

	fileWid, err := arkd.FindOrCreateWorkerIdFromFile()
	if err != nil {
		return "", err
	}

	return hn + "-" + fileWid, nil
}

func getDockerClient(ctx context.Context) (*docker.Client, error) {
	dockerClient, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, err
	}

	_, err = dockerClient.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return dockerClient, nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
