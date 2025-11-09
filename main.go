package main

import (
	"flag"
	"fmt"
	"log"
	//"os"
	"time"

	"garmin-to-ido/internal/config"
	"garmin-to-ido/internal/garmin"
	"garmin-to-ido/internal/ido"
	"garmin-to-ido/internal/sync"
)

func main() {
	// Parse command line flags
	var dateFlag, configPath string
	flag.StringVar(&dateFlag, "d", "", "Specific date to sync (format: YYYY-MM-DD). If not provided, syncs today and yesterday")
	flag.StringVar(&dateFlag, "date", "", "Specific date to sync (format: YYYY-MM-DD). If not provided, syncs today and yesterday")
	flag.StringVar(&configPath, "c", ".env", "Path to configuration file")
	flag.StringVar(&configPath, "config", ".env", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Determine dates to sync
	var datesToSync []time.Time
	if dateFlag != "" {
		date, err := time.Parse("2006-01-02", dateFlag)
		if err != nil {
			log.Fatalf("Invalid date format. Use YYYY-MM-DD: %v", err)
		}
		datesToSync = []time.Time{date}
	} else {
		// Default: today and yesterday
		now := time.Now()
		datesToSync = []time.Time{
			now,
			now.AddDate(0, 0, -1),
		}
	}

	fmt.Printf("Synchronizing bike activities for dates: ")
	for i, date := range datesToSync {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(date.Format("2006-01-02"))
	}
	fmt.Println()

	// Initialize Garmin client
	garminClient := garmin.NewClient(cfg.GarminUsername, cfg.GarminPassword)
	if err := garminClient.Login(); err != nil {
		log.Fatalf("Failed to login to Garmin: %v", err)
	}
	defer garminClient.Logout()
	fmt.Println("✓ Logged in to Garmin Connect")

	// Initialize iDO client
	idoClient, err := ido.NewClient(cfg.IdoUsername, cfg.IdoPassword)
	if err != nil {
		log.Fatalf("Failed to initialize iDO client: %v", err)
	}
	defer idoClient.Close()

	if err := idoClient.Login(); err != nil {
		log.Fatalf("Failed to login to iDO: %v", err)
	}
	fmt.Println("✓ Logged in to iDO Sport")

	// Sync activities
	syncer := sync.NewSyncer(garminClient, idoClient)
	for _, date := range datesToSync {
		fmt.Printf("\nSyncing activities for %s...\n", date.Format("2006-01-02"))
		if err := syncer.SyncBikeActivities(date); err != nil {
			log.Printf("Error syncing activities for %s: %v", date.Format("2006-01-02"), err)
			continue
		}
	}

	fmt.Println("\n✓ Synchronization completed")
}
