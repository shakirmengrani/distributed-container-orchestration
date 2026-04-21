package cmd

import (
	"context"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/shakirmengrani/distributed_docker/helper"
	"github.com/shakirmengrani/distributed_docker/types"
)

type Docker struct {
	client *client.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func NewDocker() (Docker, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	envCfg := client.WithHostFromEnv()
	cli, err := client.New(envCfg)
	return Docker{client: cli, ctx: ctx, cancel: cancelFunc}, err
}

func (docker *Docker) Create(containerConfig types.ContainerConfig) (client.ContainerCreateResult, client.ContainerStartResult, error) {
	config := &container.Config{
		Image:      containerConfig.Image,
		Env:        containerConfig.Environments,
		Volumes:    containerConfig.Volumes,
		WorkingDir: containerConfig.WorkingDir,
		Cmd:        containerConfig.Cmd,
	}
	endpoints := make(map[string]*network.EndpointSettings)
	endpoints["proxy"] = &network.EndpointSettings{}
	network := &network.NetworkingConfig{EndpointsConfig: endpoints}
	host := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
		// AutoRemove:    true,
		Resources: container.Resources{
			Memory: helper.MBToB(1024),
		},
	}
	container, err := docker.client.ContainerCreate(docker.ctx, client.ContainerCreateOptions{
		Config:           config,
		HostConfig:       host,
		NetworkingConfig: network,
		Platform:         v1.DescriptorEmptyJSON.Platform,
		Name:             containerConfig.Name,
	})
	if err != nil {
		return client.ContainerCreateResult{}, client.ContainerStartResult{}, err
	}
	start, err := docker.client.ContainerStart(docker.ctx, container.ID, client.ContainerStartOptions{})
	if err != nil {
		_, err = docker.client.ContainerRemove(docker.ctx, container.ID, client.ContainerRemoveOptions{})
		if err != nil {
			return client.ContainerCreateResult{}, client.ContainerStartResult{}, err
		}
	}
	return container, start, err
}

func (docker *Docker) Remove(containerID string) error {
	container, err := docker.client.ContainerInspect(docker.ctx, containerID, client.ContainerInspectOptions{})
	if container.Container.State.Running {
		_, err = docker.client.ContainerStop(docker.ctx, containerID, client.ContainerStopOptions{})
		if err != nil {
			return err
		}
	}
	if container.Container.HostConfig.AutoRemove != true {
		_, err = docker.client.ContainerRemove(docker.ctx, containerID, client.ContainerRemoveOptions{})
		if err != nil {
			return err
		}
	}
	return err
}

func (docker *Docker) Find(name string) (client.ContainerListResult, error) {
	filter := client.Filters{}
	if name != "" {
		filter.Add("name", name)
	}
	return docker.client.ContainerList(docker.ctx, client.ContainerListOptions{
		All:     true,
		Filters: filter,
	})
}

func (docker *Docker) Details(containerID string) (client.ContainerInspectResult, error) {
	return docker.client.ContainerInspect(docker.ctx, containerID, client.ContainerInspectOptions{})
}
