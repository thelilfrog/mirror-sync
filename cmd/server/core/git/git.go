package git

import (
	"fmt"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/storage/memory"
)

type (
	Repository struct {
		src  string
		dst  string
		auth Authentication
	}

	Authentication interface {
		Value() transport.AuthMethod
	}

	TokenAuthentication struct {
		username string
		token    string
	}

	NoAuthentication struct{}
)

func Sync(r Repository) error {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: r.src,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository from source: %w", err)
	}

	m, err := repo.CreateRemote(&config.RemoteConfig{
		Name:   "mirror",
		Mirror: true,
		URLs: []string{
			r.dst,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create remote: %w", err)
	}

	err = m.Push(&git.PushOptions{
		RemoteName: "mirror",
		Auth:       r.auth.Value(),
		RefSpecs:   []config.RefSpec{"+refs/*:refs/*"},
		Force:      true,
	})
	if err != nil {
		return fmt.Errorf("failed to push to mirror server: %w", err)
	}

	return nil
}

func (a TokenAuthentication) Value() transport.AuthMethod {
	return &http.BasicAuth{
		Username: a.username,
		Password: a.token,
	}
}

func (NoAuthentication) Value() transport.AuthMethod {
	return nil
}
