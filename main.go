package main

import (
	monitor_service "github.com/joaovwfreire/wax-monitor/cmd/monitor-service"
	rewards_service "github.com/joaovwfreire/wax-monitor/cmd/rewards-service"
)

func main() {
	/*
		// Initialize the services
		if err := monitor_service.Init(); err != nil {
			log.Fatalf("Failed to initialize monitor service: %v", err)
		}

		if err := rewards_service.Init(); err != nil {
			log.Fatalf("Failed to initialize rewards service: %v", err)
		}

		// Start the services
		if err := monitor_service.Start(); err != nil {
			log.Fatalf("Failed to start monitor service: %v", err)
		}

		if err := rewards_service.Start(); err != nil {
			log.Fatalf("Failed to start rewards service: %v", err)
		}
	*/

	monitor_service.Start()
	rewards_service.Start()
}
