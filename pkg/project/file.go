package project

import (
	"encoding/json"
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
		Source StorageSettings `yaml:"source"`
		Mirror StorageSettings `yaml:"mirror"`
	}

	StorageSettings struct {
		URL            string                   `yaml:"url"`
		Authentication AuthenticationDescriptor `yaml:"authentication"`
	}

	AuthenticationDescriptor struct {
		Basic BasicAuthenticationDescriptor `yaml:"basic"`
		Token string                        `yaml:"token"`
	}

	BasicAuthenticationDescriptor struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}
)

var (
	ErrOS      error = errors.New("failed to get os parameters")
	ErrIO      error = errors.New("failed to open file")
	ErrParsing error = errors.New("failed to parse file")
)

func LoadCurrent() (Project, error) {
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

	return decode(mainFile)
}

func LoadBytes(b []byte) (Project, error) {
	var mainFile MainFile
	if err := json.Unmarshal(b, &mainFile); err != nil {
		return Project{}, fmt.Errorf("%w: %s", ErrParsing, err)
	}

	return decode(mainFile)
}

func decode(mainFile MainFile) (Project, error) {
	wd, err := os.Getwd()
	if err != nil {
		return Project{}, fmt.Errorf("%w: cannot get current working directory path: %s", ErrOS, err)
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
		r := Repository{
			Name:        fmt.Sprintf("%s-%s", pr.Name, strings.ToLower(repoName)),
			Source:      repo.Storage.Source.URL,
			Destination: repo.Storage.Mirror.URL,
			Schedule:    repo.Schedule,
		}

		r.Authentications = make(map[string]AuthenticationSettings)
		setAuthentication(r.Authentications, "source", repo.Storage.Source.Authentication)
		setAuthentication(r.Authentications, "mirror", repo.Storage.Mirror.Authentication)

		pr.Repositories = append(pr.Repositories, r)
	}

	return pr, nil
}

func setAuthentication(m map[string]AuthenticationSettings, key string, auth AuthenticationDescriptor) {
	if len(auth.Token) > 0 {
		m[key] = AuthenticationSettings{
			Token: auth.Token,
		}
	} else if len(auth.Basic.Username) > 0 {
		m[key] = AuthenticationSettings{
			Basic: &BasicAuthenticationSettings{
				Username: auth.Basic.Username,
				Password: auth.Basic.Password,
			},
		}
	}
}

func checkConfig(mf MainFile) error {
	for _, r := range mf.Repositories {
		if len(strings.TrimSpace(r.Storage.Source.URL)) == 0 {
			return fmt.Errorf("source is empty")
		}
		if err := checkAuthenticationConfig(r.Storage.Source); err != nil {
			return err
		}
		if len(strings.TrimSpace(r.Storage.Mirror.URL)) == 0 {
			return fmt.Errorf("mirror is empty")
		}
		if err := checkAuthenticationConfig(r.Storage.Mirror); err != nil {
			return err
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

func checkAuthenticationConfig(ss StorageSettings) error {
	if len(ss.Authentication.Token) > 0 && (len(ss.Authentication.Basic.Username) > 0 || len(ss.Authentication.Basic.Password) > 0) {
		return fmt.Errorf("cannot use token and basic authentication in the same repository")
	}
	return nil
}
