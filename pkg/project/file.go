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
		Server       ServerDescriptor                `yaml:"server"`
	}

	ServerDescriptor struct {
		Hostname string `yaml:"hostname"`
		Port     int    `yaml:"port"`
		Insecure bool   `yaml:"insecure"`
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
		Name:      filepath.Base(wd),
		ServerURL: "http://localhost:8080",
	}

	if len(strings.TrimSpace(mainFile.ProjectName)) > 0 {
		pr.Name = mainFile.ProjectName
	}

	if len(strings.TrimSpace(mainFile.Server.Hostname)) > 0 {
		method := "https"
		port := 8080
		if mainFile.Server.Insecure {
			method = "http"
		}
		if mainFile.Server.Port > 0 {
			port = mainFile.Server.Port
		}
		pr.ServerURL = fmt.Sprintf("%s://%s:%d", method, mainFile.Server.Hostname, port)
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
