package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/JTGlez/gockapi/internal/manager"
)

func main() {
	configPath := flag.String("config-path", "", "Path to mock configurations directory")
	flag.Usage = printUsage
	flag.Parse()

	if *configPath == "" {
		*configPath = os.Getenv("MOCK_CONFIG_PATH")
	}
	if *configPath == "" {
		log.Fatal("config path must be provided via --config-path or MOCK_CONFIG_PATH")
	}

	if flag.NArg() < 1 {
		printUsage()
		return
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	mgr := manager.NewMockManager(*configPath)
	ctx := context.Background()

	switch cmd {
	case "start-all":
		if err := mgr.StartAll(ctx); err != nil {
			log.Fatalf("failed to start services: %v", err)
		}
		waitForSignal(mgr)
	case "start":
		if len(args) == 0 {
			log.Println("no service specified")
			return
		}
		for _, svc := range args {
			if err := mgr.StartService(ctx, svc); err != nil {
				log.Printf("failed to start %s: %v", svc, err)
			}
		}
		waitForSignal(mgr)
	case "stop-all":
		if err := mgr.StopAll(); err != nil {
			log.Fatalf("failed to stop services: %v", err)
		}
	case "stop":
		if len(args) == 0 {
			log.Println("no service specified")
			return
		}
		for _, svc := range args {
			if err := mgr.StopService(svc); err != nil {
				log.Printf("failed to stop %s: %v", svc, err)
			}
		}
	case "reload":
		if len(args) == 0 {
			log.Println("no service specified")
			return
		}
		for _, svc := range args {
			if err := mgr.ReloadService(svc); err != nil {
				log.Printf("failed to reload %s: %v", svc, err)
			}
		}
	case "status":
		status := mgr.GetStatus()
		data, _ := json.MarshalIndent(status, "", "  ")
		os.Stdout.Write(data)
		os.Stdout.Write([]byte("\n"))
	default:
		printUsage()
	}
}

func waitForSignal(mgr *manager.MockManager) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	if err := mgr.StopAll(); err != nil {
		log.Printf("failed to stop services: %v", err)
	}
}

func printUsage() {
	log.Printf(`🔧 Mock Servers CLI Tool

Usage: gockapi <command> [options]

Commands:
  start-all              Start all mock servers and wait for termination
  stop-all               Stop all running mock servers
  status                 Show status of all services

  start <service>...     Start one or more services
  stop <service>...      Stop one or more services
  reload <service>...    Reload configuration for one or more services

Options:
  --config-path string   Path to mock configurations directory (env MOCK_CONFIG_PATH)
`)

}
