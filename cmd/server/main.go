package main

import (
	"fmt"
	"log/slog"
	"mirror-sync/cmd/server/api"
	cronruntime "mirror-sync/cmd/server/core/runtime"
	"mirror-sync/cmd/server/core/storage"
	"mirror-sync/pkg/constants"
	"os"
	"runtime"
	"strconv"
)

func main() {
	dbPath := os.Getenv("MIRRORSYNC_DB_PATH")
	if len(dbPath) == 0 {
		dbPath = "/var/lib/mirror-sync/data.db"
	}

	p := os.Getenv("MIRRORSYNC_PORT")
	if len(p) == 0 {
		p = "25697"
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: bad MIRRORSYNC_PORT value: %s", err)
		os.Exit(1)
	}

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
	s := api.NewServer(data, scheduler, port)

	slog.Info("daemon listening to :" + p)
	if err := s.Server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}
}
