package dockerexecutors

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type DockerInstance struct {
	imageName string
	pathName  string
}
type ExecutionResult struct {
	Output   string
	ExitCode int
}

func New(jobId string, pathName string) DockerInstance {
	return DockerInstance{
		imageName: fmt.Sprintf("job-%v", jobId),
		pathName:  pathName,
	}
}

//not required right now since we not using dockerfile and directly using the image from the yaml

// func (d *DockerInstance) BuildImage() error {
// 	cmd := exec.Command("docker", "build", "-t", d.imageName, d.pathName)
// 	output, err := cmd.CombinedOutput()
// 	if err != nil {
// 		log.Printf("[Docker Build Error] An error occurred while building the image : %v", string(output))
// 		return err
// 	}
// 	log.Printf("Docker image build successful : %v", string(output))
// 	return nil
// }

func (d *DockerInstance) RunCommandOnJobImage(image string, command string) (ExecutionResult, error) {
	image = strings.TrimSpace(image)
	if image == "" {
		return ExecutionResult{ExitCode: -1}, errors.New("job image is required")
	}

	cmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"-v", fmt.Sprintf("%s:/workspace", d.pathName),
		"-w", "/workspace",
		image,
		"sh",
		"-c",
		command,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExecutionResult{Output: string(output), ExitCode: exitErr.ExitCode()}, nil
		}
		return ExecutionResult{Output: string(output), ExitCode: -1}, fmt.Errorf("failed to execute command on job image %q: %w", image, err)
	}
	return ExecutionResult{Output: string(output), ExitCode: 0}, nil
}
