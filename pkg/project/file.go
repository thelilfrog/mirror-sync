package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/robfig/cron/v3"
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
		Source string `yaml:"source"`
		Mirror string `yaml:"mirror"`
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

	if err := checkConfig(mainFile); err != nil {
		return Project{}, fmt.Errorf("failed to validate configuration: %w", err)
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
			Destination: repo.Storage.Mirror,
			Schedule:    repo.Schedule,
		})
	}

	return pr, nil
}

func checkConfig(mf MainFile) error {
	for _, r := range mf.Repositories {
		if len(strings.TrimSpace(r.Storage.Source)) == 0 {
			return fmt.Errorf("source is empty")
		}
		if len(strings.TrimSpace(r.Storage.Mirror)) == 0 {
			return fmt.Errorf("mirror is empty")
		}
		if len(strings.TrimSpace(r.Schedule)) == 0 {
			return fmt.Errorf("schedule is empty")
		}
		if _, err := cron.ParseStandard(r.Schedule); err != nil {
			return fmt.Errorf("failed to validate schedule: %w", err)
		}
	}

	return nil
}
