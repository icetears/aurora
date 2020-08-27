package main

import (
	"os"

	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"
	"io"
)

type DockerInputTemplate struct {
	Name       string
	Image      string
	Cmd        []string
	Env        []string
	Permitions []string
	Volume     []string
}

func dockerExec(t *DockerInputTemplate) error {
	fmt.Println("docker")
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err)
		return err
	}

	reader, err := cli.ImagePull(ctx, t.Image, types.ImagePullOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: t.Image,
		Cmd:   t.Cmd,
		Env:   t.Env,
		Tty:   true,
	}, nil, nil, t.Name)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Println(err)
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			fmt.Println(err)
			return err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		fmt.Println(err)
		return err
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}
