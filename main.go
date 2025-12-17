package main

import (
	"flag"
	"fmt"
	"linkedin-automation/config"
	"linkedin-automation/internal/auth"
	"linkedin-automation/internal/browser"
	"linkedin-automation/internal/connect"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/message"
	"linkedin-automation/internal/search"
	"linkedin-automation/internal/stealth"
	"linkedin-automation/internal/storage"
	"log"
	"os"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "./config/config.yaml", "Path to config file")
	searchQuery := flag.String("query", "", "Search query (job title)")
	searchLocation := flag.String("location", "", "Search location")
	searchCompany := flag.String("company", "", "Company name")
	maxResults := flag.Int("max", 10, "Maximum number of profiles to process")
	sendConnections := flag.Bool("connect", false, "Send connection requests")
	sendMessages := flag.Bool("message", false, "Send follow-up messages")
	flag.Parse()

	fmt.Println(`
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║           LinkedIn Automation Tool v1.0                         ║
║           ⚠️  FOR EDUCATIONAL PURPOSES ONLY ⚠️                  ║
║                                                                  ║
║   WARNING: This violates LinkedIn's Terms of Service           ║
║   Do NOT use this on production accounts                        ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
	`)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	lgr, err := logger.New(cfg.Logging.Level, cfg.Logging.File, cfg.Logging.Console)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer lgr.Close()

	lgr.Info("Starting LinkedIn Automation Tool")
	lgr.Info("Config loaded from: %s", *configPath)

	// Check business hours if enabled
	if cfg.Stealth.BusinessHoursOnly {
		if !stealth.IsBusinessHours(cfg.Stealth.WorkStartHour, cfg.Stealth.WorkEndHour) {
			lgr.Info("Outside business hours. Waiting...")
			stealth.WaitForBusinessHours(cfg.Stealth.WorkStartHour, cfg.Stealth.WorkEndHour)
		}
	}

	// Initialize storage
	store, err := storage.New(cfg.Storage.DBPath)
	if err != nil {
		lgr.Error("Failed to initialize storage: %v", err)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize browser
	lgr.Info("Initializing browser...")
	br, err := browser.New(cfg, lgr)
	if err != nil {
		lgr.Error("Failed to initialize browser: %v", err)
		os.Exit(1)
	}
	defer br.Close()

	page := br.Page()

	// Authenticate
	lgr.Info("Authenticating...")
	authenticator := auth.New(page, cfg, lgr)
	if err := authenticator.Login(); err != nil {
		lgr.Error("Authentication failed: %v", err)
		os.Exit(1)
	}

	lgr.Info("✓ Successfully authenticated")

	// Wait after login
	stealth.RandomDelay(2000, 4000)

	// Execute actions based on flags
	if *sendConnections {
		if *searchQuery == "" {
			lgr.Error("Search query is required for sending connections")
			os.Exit(1)
		}

		// Search for people
		lgr.Info("Searching for people...")
		searcher := search.New(page, cfg, lgr)
		profiles, err := searcher.SearchPeople(*searchQuery, *searchLocation, *searchCompany, *maxResults)
		if err != nil {
			lgr.Error("Search failed: %v", err)
			os.Exit(1)
		}

		lgr.Info("✓ Found %d profiles", len(profiles))

		// Send connection requests
		lgr.Info("Sending connection requests...")
		connector := connect.New(page, cfg, lgr, store)

		note := os.Getenv("CONNECTION_NOTE")
		if note == "" {
			note = "Hi {name}, I'd love to connect with you!"
		}

		if err := connector.SendConnectionRequests(profiles, note); err != nil {
			lgr.Error("Failed to send connections: %v", err)
		}

		lgr.Info("✓ Connection requests completed")
	}

	if *sendMessages {
		// Send follow-up messages
		lgr.Info("Sending follow-up messages...")
		messenger := message.New(page, cfg, lgr, store)

		msgTemplate := os.Getenv("FOLLOW_UP_MESSAGE")
		if msgTemplate == "" {
			msgTemplate = "Thanks for connecting! Looking forward to staying in touch."
		}

		if err := messenger.SendFollowUpMessages(msgTemplate); err != nil {
			lgr.Error("Failed to send messages: %v", err)
		}

		lgr.Info("✓ Follow-up messages completed")
	}

	if !*sendConnections && !*sendMessages {
		lgr.Info("No action specified. Use -connect or -message flags")
		fmt.Println(`
Usage Examples:
  # Search and send connection requests
  go run main.go -connect -query "Software Engineer" -location "San Francisco" -max 20

  # Send follow-up messages to accepted connections
  go run main.go -message

  # Combined
  go run main.go -connect -message -query "Product Manager" -company "Google" -max 10
		`)
	}

	lgr.Info("Automation completed successfully")
	fmt.Println("\n✓ All tasks completed. Check logs for details.")
}
