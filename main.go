package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/flashbots/builder-testsuite/docker"
	"github.com/flashbots/builder-testsuite/e2e"
	"gopkg.in/inconshreveable/log15.v2"
)

func main() {
	var (
		loglevel    = flag.Int("loglevel", 3, "Log `level` for system events. Supports values 0-5.")
		repo        = flag.String("repo", "", "github repo to build the builder docker image.")
		branch      = flag.String("branch", "", "branch to build the builder docker image.")
		algoType    = flag.String("algo", "greedy", "name of algo type to run the builder docker image.")
		onlyBuilder = flag.Bool("only-builder", true, "only run the builder")
	)

	// Parse the flags and configure the logger.
	flag.Parse()
	log15.Root().SetHandler(log15.LvlFilterHandler(log15.Lvl(*loglevel), log15.StreamHandler(os.Stderr, log15.TerminalFormat())))

	// Create the docker backend
	cfg := &docker.Config{
		Repo:    *repo,
		Branch:  *branch,
		BaseDir: ".",
		Logger:  log15.Root(),
	}
	dockerBackend, err := docker.NewDockerBackend(cfg)
	if err != nil {
		fatal(err)
	}

	// Set up the context for CLI interrupts.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sig
		cancel()
	}()

	// Build docker images
	if err := dockerBackend.Build(ctx, *onlyBuilder); err != nil {
		fatal(err)
	}

	// Run.
	env := e2e.TestEnv{
		AlgoType: *algoType,
	}
	runner := e2e.NewTestRunner(&env, dockerBackend)

	result, err := runner.Run(ctx)
	if err != nil {
		fatal(err)
	}
	log15.Info("test run finished", "tests", result.Tests, "failed", result.TestsFailed)

	switch result.TestsFailed {
	case 0:
	case 1:
		fatal(errors.New("1 test failed"))
	default:
		fatal(fmt.Errorf("%d tests failed", result.TestsFailed))
	}
}

func fatal(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}
