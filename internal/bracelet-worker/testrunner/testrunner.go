package testrunner

import (
	dockerexecutors "bracelet-cicd/internal/bracelet-worker/docker-executors"
	"fmt"
	"log"
)

type Result struct {
	Passed bool   `json:"passed"`
	Output string `json:"output"`
}

func Run(dockerIns *dockerexecutors.DockerInstance) (Result, error) {
	execResult, err := dockerIns.RunCommandOnImage("npm run test")
	if err != nil {
		return Result{}, fmt.Errorf("failed to run tests on image: %w", err)
	}
	result := Result{
		Passed: execResult.ExitCode == 0,
		Output: execResult.Output,
	}
	log.Printf("Test output:\n%s\n", execResult.Output)
	return result, nil
}
