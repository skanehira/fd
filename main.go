package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/ktr0731/go-fuzzyfinder"
)

var (
	up   = flag.Bool("u", false, "up container")
	down = flag.Bool("d", false, "stop container")
)

var ErrNoContainer = errors.New("not found contianer")

type Container struct {
	ID   string
	Name string
}

func main() {
	flag.Parse()

	if !*up && !*down {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	args := filters.NewArgs()

	if *up {
		for _, s := range []string{"exited", "created", "paused", "dead"} {
			args.Add("status", s)
		}
	} else if *down {
		args.Add("status", "running")
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return ErrNoContainer
	}

	var list []Container

	for _, c := range containers {
		name := c.Names[0][1:]
		id := c.ID[:10]
		list = append(list, Container{ID: id, Name: name})
	}

	ids, err := fuzzyfinder.FindMulti(
		list,
		func(i int) string {
			return fmt.Sprintf("%s: %s", list[i].ID, list[i].Name)
		},
	)
	if err != nil {
		return err
	}

	for _, i := range ids {
		id := list[i].ID
		name := list[i].Name
		if *up {
			if err := cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{}); err != nil {
				fmt.Fprintf(os.Stderr, "failed to start container: %s\n", id)
			}
			fmt.Printf("starting container: %s\n", name)
		} else if *down {
			if err := cli.ContainerStop(context.Background(), id, nil); err != nil {
				fmt.Fprintf(os.Stderr, "failed to stop container: %s\n", name)
			}
			fmt.Printf("stopping container: %s\n", name)
		}
	}

	return nil
}
