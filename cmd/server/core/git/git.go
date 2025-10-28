package git

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/storage/memory"
)

type (
	Repository struct {
		src     string
		dst     string
		srcAuth Authentication
		dstAuth Authentication
	}

	Authentication interface {
		Value() transport.AuthMethod
	}

	TokenAuthentication struct {
		token string
	}

	BasicAuthentication struct {
		username, password string
	}

	NoAuthentication struct{}
)

func NewRepository(src, dst string, srcAuth, dstAuth Authentication) Repository {
	return Repository{
		src:     src,
		dst:     dst,
		srcAuth: srcAuth,
		dstAuth: dstAuth,
	}
}

func NewTokenAuthentication(token string) TokenAuthentication {
	return TokenAuthentication{
		token: token,
	}
}

func NewBasicAuthentication(username, password string) BasicAuthentication {
	return BasicAuthentication{
		username: username,
		password: password,
	}
}

func Sync(r Repository) error {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:  r.src,
		Auth: r.srcAuth.Value(),
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
		Auth:       r.dstAuth.Value(),
		RefSpecs:   []config.RefSpec{"+refs/*:refs/*"},
		Force:      true,
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("failed to push to mirror server: %w", err)
	}

	return nil
}

func (a TokenAuthentication) Value() transport.AuthMethod {
	return &http.BasicAuth{
		Username: "git",
		Password: a.token,
	}
}

func (a BasicAuthentication) Value() transport.AuthMethod {
	return &http.BasicAuth{
		Username: a.username,
		Password: a.password,
	}
}

func (NoAuthentication) Value() transport.AuthMethod {
	return nil
}
