package main

import (
	"flag"
	"fmt"
	"log/slog"
	"mirror-sync/cmd/server/api"
	cronruntime "mirror-sync/cmd/server/core/runtime"
	"mirror-sync/cmd/server/core/storage"
	"mirror-sync/pkg/constants"
	"os"
	"runtime"
)

func main() {
	var dbPath string
	flag.StringVar(&dbPath, "db-path", "/var/lib/mirror-sync/data.db", "path to the sqlite database")
	flag.Parse()

	fmt.Printf("mirror-sync daemon -- v%s.%s.%s\n\n", constants.Version, runtime.GOOS, runtime.GOARCH)

	data, err := storage.OpenDB(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}

	if err := data.Migrate(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}

	// runtime
	prs, err := data.List()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}

	scheduler, err := cronruntime.New(prs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}

	go scheduler.Run()
	slog.Info("daemon scheduler is running")

	// api
	s := api.NewServer(data, scheduler, 8080)

	slog.Info("daemon listening to :8080")
	if err := s.Server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}
}
