package main

import (
	"context"
	"flag"
	"fmt"
	"mirror-sync/cmd/cli/commands/apply"
	"mirror-sync/cmd/cli/commands/list"
	"mirror-sync/cmd/cli/commands/remove"
	"mirror-sync/cmd/cli/commands/run"
	"mirror-sync/cmd/cli/commands/version"
	"os"

	"github.com/google/subcommands"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "fatal:", r)
		}
	}()

	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.FlagsCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")
	subcommands.Register(&version.VersionCmd{}, "help")

	subcommands.Register(&apply.ApplyCmd{}, "projects")
	subcommands.Register(&run.RunCmd{}, "projects")
	subcommands.Register(&remove.DownCmd{}, "projects")

	subcommands.Register(&list.ListCmd{}, "management")

	flag.Parse()
	ctx := context.Background()

	os.Exit(int(subcommands.Execute(ctx)))
}
