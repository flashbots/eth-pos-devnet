package e2e

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/flashbots/builder-testsuite/docker"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	resultsDir = "results"
	metricsDir = "metrics"
)

type TestID uint32

type TestSummary struct {
	Tests       int
	TestsFailed int
}

type TestResult struct {
	Pass bool
	Details string
}

type TestCase struct {
	ID 			TestID
	Name        string
	Description string
	Result 	TestResult
}

type TestEnv struct {
	AlgoType string
}

type TestRunner struct {
	env *TestEnv
	docker *docker.DockerBackend
}

func NewTestRunner(testEnv *TestEnv, docker *docker.DockerBackend) *TestRunner {
	return &TestRunner{
		env: testEnv,
		docker: docker,
	}
}

func (t *TestRunner) Run(ctx context.Context) (TestSummary, error) {
	if err := createDir(resultsDir); err != nil {
		return TestSummary{}, err
	}
	
	var testCases []TestCase

	for i, test := range getTests() {
		t.docker.Logger.Info("Running test", "name", test.Name)
		currTestCase := &TestCase{
			ID: TestID(i),
			Name: test.Name,
		}
		err := t.docker.StartLocalDevnet(ctx)
		if err != nil {
			return TestSummary{}, err
		}
		defer t.docker.StopLocalDevnet(ctx)

		builderClient, err := NewBuilderClient("127.0.0.1")
		if err != nil {
			currTestCase.Result = TestResult{
					Pass: false,
					Details: err.Error(),
				}
			continue
		}
		metrics, err := builderClient.GetMetrics()
		if err != nil {
			currTestCase.Result = TestResult{
				Pass: false,
				Details: err.Error(),
			}
			continue
		}
		
		container, err := t.docker.ContainerManager.GetBuilderContainer()
		if err != nil {
			currTestCase.Result = TestResult{
				Pass: false,
				Details: err.Error(),
			}
			continue
		}
		containerID := container.ID[:8]
		if err := writeMetricsFile(metrics, test.Name, containerID, filepath.Join(resultsDir, metricsDir)); err != nil {
			currTestCase.Result = TestResult{
				Pass: false,
				Details: err.Error(),
			}
			continue
		}
		currTestCase.Result = TestResult{
			Pass: true,
			Details: "",
		}
		testCases = append(testCases, *currTestCase)
	}

	// count the results
	var summary TestSummary
	for _, testCase := range testCases {
		summary.Tests++
		if !testCase.Result.Pass {
			summary.TestsFailed++
		}
	}
	return summary, nil
}

// writeMetricsFile writes the metric result to the results directory.
func writeMetricsFile(metrics, testName, containerId, dir string) error {
	// Create the directory if it doesn't exist.
	if err := createDir(dir); err != nil {
		return err
	}
	// metrics saved under /results/metrics/<algotype>/<test-name>-<container-id>.json
	testName = strings.ReplaceAll(strings.ToLower(testName), " ", "-")
	metricsFileName := fmt.Sprintf("%s-%s.json", testName, containerId)
	metricsFile := filepath.Join(dir, metricsFileName)
	// Write it.
	return ioutil.WriteFile(metricsFile, []byte(metrics), 0644)
}


func createDir(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			log15.Info("creating directory", "folder", dir)
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				log15.Crit("failed to create directory", "err", err)
			}
		}
		return err
	}
	if !stat.IsDir() {
		return errors.New("log output directory is a file")
	}
	return nil
}
