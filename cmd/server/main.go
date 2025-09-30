package main

import (
	"fmt"
	"mirror-sync/cmd/server/api"
	"mirror-sync/pkg/constants"
	"os"
	"runtime"
)

func main() {
	fmt.Printf("mirror-sync daemon -- v%s.%s.%s\n\n", constants.Version, runtime.GOOS, runtime.GOARCH)

	s := api.NewServer(8080)

	fmt.Println("daemon listening to :8080")
	if err := s.Server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}
}
