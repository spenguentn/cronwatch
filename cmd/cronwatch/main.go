package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/cronwatch/internal/alert"
	"github.com/example/cronwatch/internal/config"
	"github.com/example/cronwatch/internal/store"
	"github.com/example/cronwatch/internal/watcher"
)

var version = "dev"

func main() {
	cfgPath := flag.String("config", "cronwatch.yaml", "path to config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("cronwatch %s\n", version)
		os.Exit(0)
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	st, err := store.New(cfg.StorePath)
	if err != nil {
		log.Fatalf("failed to open store: %v", err)
	}

	notifier := alert.NewLogNotifier(os.Stderr)
	dispatcher := alert.NewDispatcher(notifier)
	updater := store.NewUpdater(st)

	w := watcher.New(cfg, updater, dispatcher)

	ctx, stop := signal.NotifyContext(nil, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("cronwatch %s starting, monitoring %d job(s)", version, len(cfg.Jobs))
	w.Run(ctx)
	log.Println("cronwatch stopped")
}
