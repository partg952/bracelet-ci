package parser

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

//version : "1"
//build:
//	image:golang-xx
//  jobs:
//		- run : go build
//test:
//	image:golang-xx
//	jobs:
//	    - run : go test .

type Yaml struct {
	Version string         `yaml:"version"`
	Jobs    map[string]Job `yaml:"jobs"`
}

type Job struct {
	Image string `yaml:"image"`
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Run string `yaml:"run"`
}

func ingestFile(path string) (string, error) {
	contents, err := os.ReadFile(path + "/.braceletci.yml")
	if err != nil {
		return "", fmt.Errorf("[YAML Parser Error] Error occurred while reading file : %v", err)
	}
	return string(contents), nil
}

func ParseYaml(path string) (Yaml, error) {
	contents, err := ingestFile(path)
	if err != nil {
		return Yaml{}, err
	}
	var parsedYaml Yaml
	err = yaml.Unmarshal([]byte(contents), &parsedYaml)
	if err != nil {
		return Yaml{}, err
	}
	return parsedYaml, nil
}
