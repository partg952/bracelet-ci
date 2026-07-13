package testrunner

import (
	dockerexecutors "bracelet-cicd/internal/bracelet-worker/docker-executors"
	"bracelet-cicd/internal/bracelet-worker/parser"
	"fmt"
	"log"
	"sort"
	"strings"
)

type Result struct {
	Passed bool   `json:"passed"`
	Output string `json:"output"`
}

func Run(dockerIns *dockerexecutors.DockerInstance, parsedYaml parser.Yaml) (Result, error) {
	if len(parsedYaml.Jobs) == 0 {
		return Result{}, fmt.Errorf("no jobs found in .braceletci.yml")
	}

	var output strings.Builder
	jobNames := make([]string, 0, len(parsedYaml.Jobs))
	for jobName := range parsedYaml.Jobs {
		jobNames = append(jobNames, jobName)
	}
	sort.Strings(jobNames)

	for _, jobName := range jobNames {
		job := parsedYaml.Jobs[jobName]
		if strings.TrimSpace(job.Image) == "" {
			return Result{}, fmt.Errorf("job %q has no image", jobName)
		}
		if len(job.Steps) == 0 {
			return Result{}, fmt.Errorf("job %q has no steps", jobName)
		}

		var jobCommand strings.Builder
		for stepIndex, step := range job.Steps {
			if strings.TrimSpace(step.Run) == "" {
				return Result{}, fmt.Errorf("job %q step %d has no run command", jobName, stepIndex+1)
			}

			output.WriteString(fmt.Sprintf("[%s] %s\n$ %s\n", jobName, job.Image, step.Run))
			jobCommand.WriteString(step.Run)
			jobCommand.WriteString("\n")
		}

		execResult, err := dockerIns.RunCommandOnJobImage(job.Image, "set -e\n"+jobCommand.String())
		output.WriteString(execResult.Output)
		if err != nil {
			return Result{}, fmt.Errorf("failed to run job %q: %w", jobName, err)
		}
		if execResult.ExitCode != 0 {
			result := Result{
				Passed: false,
				Output: output.String(),
			}
			log.Printf("Job output:\n%s\n", result.Output)
			return result, nil
		}
	}

	result := Result{
		Passed: true,
		Output: output.String(),
	}
	log.Printf("Job output:\n%s\n", result.Output)
	return result, nil
}
