package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type (
	MainFile struct {
		Repositories map[string]RepositoryDescriptor `yaml:"repositories"`
		ProjectName  string                          `yaml:"project_name"`
	}

	RepositoryDescriptor struct {
		Storage  GitStorage `yaml:"storage"`
		Schedule string     `yaml:"schedule"`
	}

	GitStorage struct {
		Source      string `yaml:"source"`
		Destination string `yaml:"destination"`
	}
)

var (
	ErrOS      error = errors.New("failed to get os parameters")
	ErrIO      error = errors.New("failed to open file")
	ErrParsing error = errors.New("failed to parse file")
)

func LoadCurrent() (Project, error) {
	wd, err := os.Getwd()
	if err != nil {
		return Project{}, fmt.Errorf("%w: cannot get current working directory path: %s", ErrOS, err)
	}

	f, err := os.OpenFile("./git-compose.yaml", os.O_RDONLY, 0)
	if err != nil {
		return Project{}, fmt.Errorf("%w: %s", ErrIO, err)
	}
	defer f.Close()

	var mainFile MainFile
	d := yaml.NewDecoder(f)
	if err := d.Decode(&mainFile); err != nil {
		return Project{}, fmt.Errorf("%w: %s", ErrParsing, err)
	}

	pr := Project{
		Name: filepath.Base(wd),
	}

	if len(strings.TrimSpace(mainFile.ProjectName)) > 0 {
		pr.Name = mainFile.ProjectName
	}

	for repoName, repo := range mainFile.Repositories {
		pr.Repositories = append(pr.Repositories, Repository{
			Name:        fmt.Sprintf("%s-%s", pr.Name, strings.ToLower(repoName)),
			Source:      repo.Storage.Source,
			Destination: repo.Storage.Destination,
			Schedule:    repo.Schedule,
		})
	}

	return pr, nil
}
