package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	docker "github.com/docker/docker/client"
	bolt "go.etcd.io/bbolt"

	"github.com/dkimot/ark/services/arkd/internal/api"
	"github.com/dkimot/ark/services/arkd/internal/arkd"
	"github.com/dkimot/ark/services/arkd/internal/config"
	"github.com/dkimot/ark/services/arkd/internal/orca"
	"github.com/rs/zerolog"
)

var (
  ApiVersion = "alpha"
)

func run(ctx context.Context, args []string, getenv func(string) string, stdin io.Reader, stdout, stderr io.Writer) error {
  l := zerolog.New(stdout)

  wid, err := getWorkerId(getenv("ARKD_WORKER_ID"))
  if err != nil {
    return err
  }

  cfg := config.NewConfig(config.WithApiVersion(ApiVersion), config.WithWorkerId(wid))

  db, err := bolt.Open("arkd.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
  if err != nil {
    return err
  }
  defer db.Close()

  taskStore, err := arkd.NewTaskStore(db, l)
  if err != nil {
    return err
  }

  dockerClient, err := getDockerClient(ctx)
  if err != nil {
    return err
  }
  defer dockerClient.Close()

  or, err := orca.Start(cfg, dockerClient, taskStore)
  if err != nil {
    return err
  }

  return api.StartHttpServer(
    l, 
    cfg,
    db, 
    taskStore,
    dockerClient,
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
