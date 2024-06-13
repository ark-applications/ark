package main

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func main() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer apiClient.Close()

	body, err := apiClient.ImagePull(context.Background(), "postgres:10", image.PullOptions{})
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, body)
}
