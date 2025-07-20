package main

import "log"

func main() {

	// This should be the entry point for the CLI tool
	printUsage()

}

func printUsage() {
	log.Printf(`🔧 Mock Servers CLI Tool

	Usage: %s <command> [options]

	Commands:
	start-all              Start all mock servers and wait for termination
	stop-all               Stop all running mock servers  
	status                 Show status of all services
	
	start <service>        Start a specific service
	stop <service>         Stop a specific service
	reload <service>       Reload configuration for a specific service

	Environment Variables:
	MOCK_CONFIG_PATH      Path to mock configurations directory

	`)

}
