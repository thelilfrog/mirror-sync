package version

import (
	"context"
	"flag"
	"fmt"
	"mirror-sync/cmd/cli/config"
	"mirror-sync/pkg/client"
	"mirror-sync/pkg/constants"
	"os"
	"runtime"
	"strconv"

	"github.com/google/subcommands"
)

type (
	VersionCmd struct {
	}
)

func (*VersionCmd) Name() string     { return "version" }
func (*VersionCmd) Synopsis() string { return "show version and system information" }
func (*VersionCmd) Usage() string {
	return `Usage: mirror-sync version

Print the version of the software

Options:
`
}

func (p *VersionCmd) SetFlags(f *flag.FlagSet) {
}

func (p *VersionCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	fmt.Println("Client: mirror-sync cli")
	fmt.Println(" Version:       " + constants.Version)
	fmt.Println(" API version:   " + strconv.Itoa(constants.ApiVersion))
	fmt.Println(" Go version:    " + runtime.Version())
	fmt.Println(" OS/Arch:       " + runtime.GOOS + "/" + runtime.GOARCH)

	clientConfig := config.Load()

	cli := client.New(clientConfig.Deamon.URL)
	systemInfoDaemon, err := cli.Version()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return subcommands.ExitFailure
	}

	fmt.Println()
	fmt.Println("Daemon:")
	fmt.Println(" Version:       " + systemInfoDaemon.Version)
	fmt.Println(" API version:   " + strconv.Itoa(systemInfoDaemon.APIVersion))
	fmt.Println(" Go version:    " + systemInfoDaemon.GoVersion)
	fmt.Println(" OS/Arch:       " + systemInfoDaemon.OSName + "/" + systemInfoDaemon.OSArchitecture)

	return subcommands.ExitSuccess
}

func local() {

}
