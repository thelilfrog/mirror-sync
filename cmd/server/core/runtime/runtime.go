package runtime

import (
	"fmt"
	"log/slog"
	"mirror-sync/cmd/server/core/git"
	"mirror-sync/pkg/project"

	"github.com/robfig/cron/v3"
)

type (
	Scheduler struct {
		cr  *cron.Cron
		ids map[string]map[string]cron.EntryID
	}
)

func New(prs []project.Project) (*Scheduler, error) {
	s := &Scheduler{
		cr:  cron.New(),
		ids: make(map[string]map[string]cron.EntryID),
	}

	for _, pr := range prs {
		if err := s.Add(pr); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Scheduler) Add(pr project.Project) error {
	s.ids[pr.Name] = make(map[string]cron.EntryID)
	for _, repo := range pr.Repositories {
		var srcAuth git.Authentication = git.NoAuthentication{}
		var dstAuth git.Authentication = git.NoAuthentication{}
		if v, ok := repo.Authentications["source"]; ok {
			if len(v.Token) > 0 {
				srcAuth = git.NewTokenAuthentication(v.Token)
			} else if v.Basic != nil {
				srcAuth = git.NewBasicAuthentication(v.Basic.Username, v.Basic.Password)
			}
		}
		if v, ok := repo.Authentications["mirror"]; ok {
			if len(v.Token) > 0 {
				dstAuth = git.NewTokenAuthentication(v.Token)
			} else if v.Basic != nil {
				dstAuth = git.NewBasicAuthentication(v.Basic.Username, v.Basic.Password)
			}
		}
		r := git.NewRepository(repo.Source, repo.Destination, srcAuth, dstAuth)
		id, err := s.cr.AddFunc(repo.Schedule, func() {
			slog.Info(fmt.Sprintf("[%s] starting sync...", repo.Name))
			if err := git.Sync(r); err != nil {
				slog.Error(fmt.Sprintf("[%s] failed to sync repository: %s", repo.Name, err))
				return
			}
			slog.Info(fmt.Sprintf("[%s] synced", repo.Name))
		})
		if err != nil {
			return err
		}
		s.ids[pr.Name][repo.Name] = id

		slog.Info(fmt.Sprintf("[%s] scheduled with '%s'", repo.Name, repo.Schedule))
	}

	return nil
}

func (s *Scheduler) Remove(pr project.Project) {
	if v, ok := s.ids[pr.Name]; ok {
		for name, id := range v {
			s.cr.Remove(id)
			slog.Info(fmt.Sprintf("[%s] remove from being run in the future.", name))
		}
	}
	delete(s.ids, pr.Name)
}

// Run the cron scheduler, or no-op if already running.
func (s *Scheduler) Run() {
	s.cr.Run()
}
