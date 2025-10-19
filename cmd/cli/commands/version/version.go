package version

import (
	"context"
	"flag"
	"fmt"
	"mirror-sync/pkg/constants"
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
	local()
	return subcommands.ExitSuccess
}

func local() {
	fmt.Println("Client: mirror-sync cli")
	fmt.Println(" Version:       " + constants.Version)
	fmt.Println(" API version:   " + strconv.Itoa(constants.ApiVersion))
	fmt.Println(" Go version:    " + runtime.Version())
	fmt.Println(" OS/Arch:       " + runtime.GOOS + "/" + runtime.GOARCH)
}
