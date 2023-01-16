package docker

import (
	"fmt"

	"github.com/docker/compose/v2/pkg/api"
	moby "github.com/docker/docker/api/types"
	docker "github.com/fsouza/go-dockerclient"
	"gopkg.in/inconshreveable/log15.v2"
)

type ContainerOptions struct {
	Cmd     []string
	LogFile string
}

type ContainerManager struct {
	client     *docker.Client
	config     *Config
	logger     log15.Logger
	containers []moby.Container
}

func NewContainerManager(client *docker.Client, config *Config) *ContainerManager {
	return &ContainerManager{
		client: client,
		config: config,
		logger: config.Logger,
	}
}

func (c *ContainerManager) AddContainers(container ...moby.Container) {
	c.containers = append(c.containers, container...)
}

func (c *ContainerManager) GetBuilderContainer() (moby.Container, error) {
	for _, container := range c.containers {
		if container.Labels[api.ServiceLabel] == "geth" {
			return container, nil
		}
	}
	return moby.Container{}, fmt.Errorf("no builder container found")
}