package repository

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func Clone(repoURL string, commitSHA string) (string, string, error) {
	if repoURL == "" {
		return "", "", errors.New("clone URL is empty")
	}
	if commitSHA == "" {
		return "", "", errors.New("commit SHA is empty")
	}

	pathName, err := os.MkdirTemp("", "")
	if err != nil {
		return "", "", fmt.Errorf("failed to create tmp directory: %w", err)
	}
	log.Printf("Random generated folder: %s\n", pathName)

	var cloneOutput strings.Builder

	cmd := exec.Command("git", "clone", repoURL, pathName)
	output, err := cmd.CombinedOutput()
	cloneOutput.WriteString(string(output))
	if err != nil {
		os.RemoveAll(pathName)
		return "", cloneOutput.String(), fmt.Errorf("git clone failed: %w: %s", err, output)
	}

	cmd = exec.Command("git", "checkout", commitSHA)
	cmd.Dir = pathName
	output, err = cmd.CombinedOutput()
	cloneOutput.WriteString(string(output))
	if err != nil {
		os.RemoveAll(pathName)
		return "", cloneOutput.String(), fmt.Errorf("git checkout failed: %w: %s", err, output)
	}

	return pathName, cloneOutput.String(), nil
}
