package main

import (
	"flag"
	"fmt"
	"log/slog"
	"mirror-sync/cmd/server/api"
	"mirror-sync/cmd/server/core/config"
	cronruntime "mirror-sync/cmd/server/core/runtime"
	"mirror-sync/cmd/server/core/storage"
	"mirror-sync/pkg/constants"
	"os"
	"runtime"
	"strconv"
)

func main() {
	ecpath := os.Getenv("MIRRORSYNC_CONFIG_PATH")
	if len(ecpath) == 0 {
		ecpath = "/etc/mirror-sync"
	}

	var configPath string
	flag.StringVar(&configPath, "config", ecpath, "path to the configuration folder")
	flag.Parse()

	c, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}

	dbPath := os.Getenv("MIRRORSYNC_DB_PATH")
	if len(dbPath) == 0 {
		dbPath = c.Database.Path
	}

	p := os.Getenv("MIRRORSYNC_PORT")
	if len(p) == 0 {
		p = fmt.Sprintf("%d", c.Server.Port)
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
	s := api.NewServer(data, scheduler, c.Server.Address, port)

	slog.Info(fmt.Sprintf("daemon listening to %s:%s", c.Server.Address, p))
	if err := s.Server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start server:", err.Error())
		os.Exit(1)
	}
}
