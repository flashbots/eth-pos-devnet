package docker

import (
	"context"
	"fmt"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
	"gopkg.in/inconshreveable/log15.v2"
)

// BuilderConfig is the configuration for the parameters of the builder
type BuilderConfig struct {
	Algotype         string
	MaxMergedBundles int
	Recommit         string
}

// Config is the configuration of the docker backend.
type Config struct {
	Repo          string
	Branch        string
	BaseDir       string
	BuilderConfig BuilderConfig
	Logger        log15.Logger
}

type DockerBackend struct {
	Builder          *Builder
	ContainerManager *ContainerManager
	ComposeService   *ComposeService
	Config           *Config
	Logger           log15.Logger
}

func NewDockerBackend(config *Config) (*DockerBackend, error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("error creating docker client %s", err)
	}
	compose, err := NewComposeService(filepath.Join(config.BaseDir, "clients", "docker-compose.yml"))
	if err != nil {
		return nil, fmt.Errorf("error creating docker compose %s", err)
	}

	return &DockerBackend{
		Builder:          NewBuilder(client, config),
		ContainerManager: NewContainerManager(client, config),
		ComposeService:   compose,
		Config:           config,
		Logger:           config.Logger,
	}, nil
}

func (d *DockerBackend) Build(ctx context.Context, onlyBuilder bool) error {
	err := d.Builder.BuildGeth(ctx)
	if err != nil {
		return err
	}
	if !onlyBuilder {
		err = d.Builder.BuildPrysm(ctx)
		if err != nil {
			return err
		}
		err = d.ComposeService.Build(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DockerBackend) StartBuilder(ctx context.Context) error {
	return nil
}

func (d *DockerBackend) StartLocalDevnet(ctx context.Context) error {
	err := d.ComposeService.Up(ctx)
	if err != nil {
		return err
	}
	containers, err := d.ComposeService.GetContainers(ctx)
	if err != nil {
		return err
	}
	d.ContainerManager.AddContainers(containers...)
	return nil
}

func (d *DockerBackend) StopLocalDevnet(ctx context.Context) error {
	err := d.ComposeService.Down(ctx)
	if err != nil {
		return err
	}
	return nil
}