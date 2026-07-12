package dockerexecutors

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
)

type DockerInstance struct {
	imageName string
	pathName string

}

type ExecutionResult struct {
	Output   string
	ExitCode int
}
func New(jobId string , pathName string) DockerInstance {
	return DockerInstance{
		imageName: fmt.Sprintf("job-%v" , jobId),
		pathName: pathName,

	}
}

func (d *DockerInstance) BuildImage() error {
	cmd := exec.Command("docker", "build", "-t", d.imageName ,d.pathName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[Docker Build Error] An error occurred while building the image : %v" , string(output))
		return err
	}
	log.Printf("Docker image build successful : %v", string(output))
	return nil
}

func (d *DockerInstance) RunCommandOnImage(command string) (ExecutionResult, error) {
	cmd := exec.Command("docker", "run", "--rm", d.imageName, "sh", "-c", command)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExecutionResult{Output: string(output), ExitCode: exitErr.ExitCode()}, nil
		}
		return ExecutionResult{Output: string(output), ExitCode: -1}, fmt.Errorf("failed to execute command on image: %w", err)
	}
	return ExecutionResult{Output: string(output), ExitCode: 0}, nil
}


