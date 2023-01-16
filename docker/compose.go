package docker

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	moby "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"gopkg.in/inconshreveable/log15.v2"
)

type ComposeService struct {
	cli     *command.DockerCli
	service *api.ServiceProxy
	project *types.Project
	logger  log15.Logger
}

func NewComposeService(composeFile string) (*ComposeService, error) {
	dockercli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("error creating client %s", err)
	}

	err = dockercli.Initialize(flags.NewClientOptions())
	if err != nil {
		return nil, fmt.Errorf("error initing docker client %s", err)
	}

	proxy := api.NewServiceProxy()
	service := proxy.WithService(compose.NewComposeService(dockercli))
	projectOpts, err := cli.NewProjectOptions([]string{composeFile})
	if err != nil {
		return nil, fmt.Errorf("error creating project options %s", err)
	}

	project, err := cli.ProjectFromOptions(projectOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating project %s", err)
	}

	for i, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False", // default
		}
		project.Services[i] = s
	}

	return &ComposeService{
		cli:     dockercli,
		service: service,
		project: project,
		logger:  log15.Root(),
	}, nil
}

func (c *ComposeService) GetContainers(ctx context.Context) ([]moby.Container, error) {
	var containers []moby.Container
	containers, err := c.cli.Client().ContainerList(ctx, moby.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("label", fmt.Sprintf("%s=%s", api.ProjectLabel, c.project.Name))),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing containers %s", err)
	}
	return containers, nil
}

func (c *ComposeService) Build(ctx context.Context) error {
	return c.service.Build(ctx, c.project, api.BuildOptions{Quiet: true})
}

func (c *ComposeService) Up(ctx context.Context) error {
	upFunc := func(ctx context.Context) error {
		return c.service.Up(ctx, c.project, api.UpOptions{})
	}
	return wrapStderr(ctx, upFunc)
}

func (c *ComposeService) Down(ctx context.Context) error {
	downFunc := func(ctx context.Context) error {
		opts := api.DownOptions{RemoveOrphans: true, Project: c.project, Images: "local"}
		return c.service.Down(ctx, c.project.Name, opts)
	}
	return wrapStderr(ctx, downFunc)
}

func wrapStderr(ctx context.Context, f func(context.Context) error) error {
	capturedOutput := os.Stderr
	_, w, err := os.Pipe()
	if err != nil {
		return err
	}
	os.Stderr = w

	err = f(ctx)
	if err != nil {
		return err
	}

	w.Close()
	os.Stderr = capturedOutput
	return nil
}