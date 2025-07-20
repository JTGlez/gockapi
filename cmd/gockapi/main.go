package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"path/filepath"

	"github.com/JTGlez/gockapi/internal/manager"
	"github.com/JTGlez/gockapi/internal/server/process_killer"
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
	mProcessKiller := process_killer.NewLinuxProcessKiller()

	switch cmd {
	case "start-all":
		if err := mgr.StartAll(ctx); err != nil {
			log.Fatalf("failed to start services: %v", err)
		}
		time.Sleep(1 * time.Second)
		log.Println("🚀 All mock servers are up and running!")
		waitForSignal(mgr)
	case "start":
		if len(args) == 0 {
			log.Println("no service specified")
			return
		}
		for _, svc := range args {
			if err := mgr.StartService(ctx, svc); err != nil {
				log.Printf("❌ Failed to start %s: %v", svc, err)
			} else {
				log.Printf("✅ Service %s started successfully", svc)
			}
		}
		time.Sleep(1 * time.Second)
		log.Println("🚀 All requested mock servers are up and running!")
		waitForSignal(mgr)
	case "stop-all":
		// Improved logic: statelessly stop all services by scanning config directory
		files, err := filepath.Glob(filepath.Join(*configPath, "*.json"))
		if err != nil {
			log.Fatalf("failed to list config files: %v", err)
		}
		if len(files) == 0 {
			log.Println("No service configs found to stop.")
			return
		}
		numKilled := 0
		errors := []string{}
		for _, file := range files {
			serviceName := filepath.Base(file)
			serviceName = serviceName[:len(serviceName)-len(filepath.Ext(serviceName))]
			cfg, cfgErr := mgr.GetConfigReader().ReadServiceConfig(serviceName)
			if cfgErr != nil {
				errors = append(errors, "❌ Could not read config for "+serviceName+": "+cfgErr.Error())
				continue
			}
			killErr := mProcessKiller.KillProcessOnPort(cfg.Port)
			if killErr != nil {
				if numKilled == 0 || (killErr.Error() != "could not find process on port "+fmt.Sprint(cfg.Port)+": ") {
					errors = append(errors, "❌ Could not kill process for "+serviceName+" on port "+fmt.Sprint(cfg.Port)+": "+killErr.Error())
				}
			} else {
				numKilled++
			}
		}
		for _, e := range errors {
			log.Println(e)
		}
		if numKilled > 0 {
			log.Println("✅ All detected mock servers have been stopped.")
		} else if len(errors) > 0 {
			log.Println("No running mock servers were found to stop.")
		}
		return
	case "stop":
		if len(args) == 0 {
			log.Println("no service specified")
			return
		}
		for _, svc := range args {
			if err := mgr.StopService(svc); err != nil {
				cfg, cfgErr := mgr.GetConfigReader().ReadServiceConfig(svc)
				if cfgErr != nil {
					log.Printf("❌ Could not read config for %s: %v", svc, cfgErr)
					continue
				}
				if killErr := mProcessKiller.KillProcessOnPort(cfg.Port); killErr != nil {
					log.Printf("❌ Could not kill process for %s on port %d: %v", svc, cfg.Port, killErr)
				} else {
					log.Printf("✅ Successfully killed process for %s on port %d", svc, cfg.Port)
				}
			} else {
				log.Printf("✅ Service %s stopped successfully via manager", svc)
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
		files, err := filepath.Glob(filepath.Join(*configPath, "*.json"))
		if err != nil {
			log.Fatalf("failed to list config files: %v", err)
		}
		if len(files) == 0 {
			log.Println("No service configs found.")
			return
		}
		running := []string{}
		for _, file := range files {
			serviceName := filepath.Base(file)
			serviceName = serviceName[:len(serviceName)-len(filepath.Ext(serviceName))]
			cfg, cfgErr := mgr.GetConfigReader().ReadServiceConfig(serviceName)
			if cfgErr != nil {
				continue
			}
			address := net.JoinHostPort("localhost", fmt.Sprint(cfg.Port))
			conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
			if err == nil {
				running = append(running, fmt.Sprintf("%s (port %d)", serviceName, cfg.Port))
				conn.Close()
			}
		}
		if len(running) == 0 {
			log.Println("No running mock servers found.")
			return
		}
		log.Println("Running mock servers:")
		for _, s := range running {
			log.Println("  -", s)
		}
	default:
		printUsage()
	}
}

func waitForSignal(mgr *manager.MockManager) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	if err := mgr.StopAll(); err != nil {
		if err.Error() != "mock manager is not running" {
			log.Printf("failed to stop services: %v", err)
		}
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
