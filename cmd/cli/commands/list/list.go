package list

import (
	"context"
	"flag"
	"fmt"
	"mirror-sync/cmd/cli/config"
	"mirror-sync/pkg/client"
	"mirror-sync/pkg/project"
	"os"

	"github.com/google/subcommands"
)

type (
	ListCmd struct {
		projectName string
	}
)

func (*ListCmd) Name() string     { return "list" }
func (*ListCmd) Synopsis() string { return "list the scheduled projects" }
func (*ListCmd) Usage() string {
	return `Usage: mirror-sync list [--project-name]

list the scheduled projects

Options:
`
}

func (p *ListCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.projectName, "project-name", "", "show only one project")
}

func (p *ListCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	clientConfig := config.Load()

	cli := client.New(clientConfig.Deamon.URL)

	prs, err := cli.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	for _, pr := range prs {
		print(pr)
	}

	return subcommands.ExitSuccess
}

func print(pr project.Project) {
	fmt.Println(pr.Name)
	fmt.Println("------------------")

	for _, repo := range pr.Repositories {
		fmt.Printf("%s | %-20s | %s -> %s | %s\n", repo.UUID, repo.Name, repo.Source, repo.Destination, repo.Schedule)
	}
}
