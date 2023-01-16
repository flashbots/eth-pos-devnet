package docker

import (
	"context"
	"io"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
	"gopkg.in/inconshreveable/log15.v2"
)

// Builder takes care of building docker images.
type Builder struct {
	client *docker.Client
	config *Config
	logger log15.Logger
}

func NewBuilder(client *docker.Client, config *Config) *Builder {
	return &Builder{
		client: client,
		config: config,
		logger: config.Logger,
	}
}

func (b *Builder) BuildGeth(ctx context.Context) error {
	contextDir := filepath.Join(b.config.BaseDir, "clients/execution")
	imageTag := "flashbots/builder:latest"
	return b.buildImage(ctx, contextDir, "Dockerfile", b.config.Branch, b.config.Repo, imageTag)
}

func (b *Builder) BuildPrysm(ctx context.Context) error {
	contextDir := filepath.Join(b.config.BaseDir, "clients/consensus")
	imageTag := "flashbots/prysm/beacon-chain:latest"
	return b.buildImage(ctx, contextDir, "Dockerfile", "", "", imageTag)
}

// buildImage builds a single docker image from the specified context.
// branch specifes a build argument to use a specific github source branch.
func (b *Builder) buildImage(ctx context.Context, contextDir, dockerFile, branch, repo, imageTag string) error {
	logger := b.logger.New("image", imageTag)
	context, err := filepath.Abs(contextDir)
	if err != nil {
		logger.Error("can't find path to context directory", "err", err)
		return err
	}

	opts := docker.BuildImageOptions{
		Context:      ctx,
		Name:         imageTag,
		OutputStream: io.Discard,
		ContextDir:   context,
		Dockerfile:   dockerFile,
	}
	logctx := []interface{}{"dir", contextDir}
	var buildArgs []docker.BuildArg
	if branch != "" {
		buildArgs = append(buildArgs, docker.BuildArg{
			Name:  "branch",
			Value: branch,
		})
		logctx = append(logctx, "branch", branch)
	}
	if repo != "" {
		buildArgs = append(buildArgs, docker.BuildArg{
			Name:  "repo",
			Value: repo,
		})
		logctx = append(logctx, "repo", repo)
	}
	opts.BuildArgs = buildArgs

	logger.Info("building image", logctx...)
	if err := b.client.BuildImage(opts); err != nil {
		logger.Error("image build failed", "err", err)
		return err
	}
	return nil
}
