package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"

	"github.com/dkimot/ark/arkcluster/internal/api"
	"github.com/dkimot/ark/arkcluster/internal/config"
)

var (
	ApiVersion = "alpha"
)

func run(ctx context.Context, _ []string, getenv func(string) string, _ io.Reader, stdout, _ io.Writer) (err error) {
  l := zerolog.New(stdout).With().
    Timestamp().
    Str("role", "ark-cluster").
    Logger()

  cfg := config.NewConfig(config.WithApiVersion(ApiVersion))

  // open DB

  // set up dependencies

  // start api
  return api.StartHttpServer(l, cfg)
}

func main() {
  ctx := context.Background()
  if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err)
    os.Exit(1)
  }
}
