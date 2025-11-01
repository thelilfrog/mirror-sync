package run

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
	RunCmd struct {
	}
)

func (*RunCmd) Name() string     { return "run" }
func (*RunCmd) Synopsis() string { return "run the current project schedule" }
func (*RunCmd) Usage() string {
	return `Usage: mirror-sync run

run the current project
`
}

func (p *RunCmd) SetFlags(f *flag.FlagSet) {
}

func (p *RunCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	clientConfig := config.Load()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	defaultValues := project.DefaultValues{
		DaemonURL:   clientConfig.Deamon.URL,
		ProjectName: filepath.Base(wd),
	}

	projectConfig, err := project.LoadCurrent(defaultValues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	cli := client.New(projectConfig.ServerURL)
	if err := cli.RunOne(projectConfig); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
