package remove

import (
	"context"
	"flag"
	"fmt"
	"mirror-sync/cmd/cli/config"
	"mirror-sync/pkg/client"
	"mirror-sync/pkg/project"
	"os"
	"path/filepath"

	"github.com/google/subcommands"
)

type (
	DownCmd struct {
		projectName string
	}
)

func (*DownCmd) Name() string     { return "down" }
func (*DownCmd) Synopsis() string { return "remove the current project schedule" }
func (*DownCmd) Usage() string {
	return `Usage: mirror-sync down

remove the current project
`
}

func (p *DownCmd) SetFlags(f *flag.FlagSet) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	f.StringVar(&p.projectName, "project-name", filepath.Base(wd), "set the project name")
}

func (p *DownCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	clientConfig := config.Load()

	defaultValues := project.DefaultValues{
		DaemonURL:   clientConfig.Deamon.URL,
		ProjectName: p.projectName,
	}

	projectConfig, err := project.LoadCurrent(defaultValues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	cli := client.New(projectConfig.ServerURL)
	if err := cli.Remove(projectConfig); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
