package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/dkimot/ark/arkcluster/internal/api"
	"github.com/dkimot/ark/arkcluster/internal/config"
	"github.com/dkimot/ark/arkcluster/internal/models"
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
  db, err := gorm.Open(sqlite.Open("ark.db"), &gorm.Config{})
  if err != nil {
    return err
  }

  if err := db.AutoMigrate(&models.Stack{}); err != nil {
    return fmt.Errorf("could not automigrate stacks: %w", err)
  }
  if err := db.AutoMigrate(&models.Deployment{}); err != nil {
    return fmt.Errorf("could not automigrate deployments: %w", err)
  }

  // set up dependencies

  // start api
  return api.StartHttpServer(l, cfg, db)
}

func main() {
  ctx := context.Background()
  if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
    fmt.Fprintf(os.Stderr, "%s\n", err)
    os.Exit(1)
  }
}
