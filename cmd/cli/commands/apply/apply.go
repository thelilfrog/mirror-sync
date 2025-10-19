package apply

import (
	"context"
	"flag"
	"fmt"
	"mirror-sync/pkg/client"
	"mirror-sync/pkg/project"
	"os"

	"github.com/google/subcommands"
)

type (
	ApplyCmd struct {
	}
)

func (*ApplyCmd) Name() string     { return "apply" }
func (*ApplyCmd) Synopsis() string { return "apply the current project settings" }
func (*ApplyCmd) Usage() string {
	return `Usage: mirror-sync apply

apply the current project settings

Options:
`
}

func (p *ApplyCmd) SetFlags(f *flag.FlagSet) {
}

func (p *ApplyCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	projectConfig, err := project.LoadCurrent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	cli := client.New(projectConfig.ServerURL)
	if err := cli.Apply(projectConfig); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
