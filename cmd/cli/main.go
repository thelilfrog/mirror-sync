package main

import (
	"context"
	"flag"
	"mirror-sync/cmd/cli/commands/apply"
	"mirror-sync/cmd/cli/commands/version"
	"os"

	"github.com/google/subcommands"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "help")
	subcommands.Register(subcommands.FlagsCommand(), "help")
	subcommands.Register(subcommands.CommandsCommand(), "help")
	subcommands.Register(&version.VersionCmd{}, "help")

	subcommands.Register(&apply.ApplyCmd{}, "projects")

	flag.Parse()
	ctx := context.Background()

	os.Exit(int(subcommands.Execute(ctx)))
}
