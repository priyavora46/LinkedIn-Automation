package connect

import (
	"errors"
	"fmt"
	"linkedin-automation/config"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/search"
	"linkedin-automation/internal/stealth"
	"linkedin-automation/internal/storage"
	"strings"

	"github.com/go-rod/rod"
)

type Connector struct {
	page   *rod.Page
	cfg    *config.Config
	logger *logger.Logger
	store  *storage.Store
}

func New(page *rod.Page, cfg *config.Config, log *logger.Logger, store *storage.Store) *Connector {
	return &Connector{
		page:   page,
		cfg:    cfg,
		logger: log,
		store:  store,
	}
}

func (c *Connector) SendConnectionRequests(profiles []search.Profile, note string) error {
	c.logger.Info("Starting to send connection requests to %d profiles", len(profiles))

	// Check today's limit
	todayCount, err := c.store.GetConnectionsCountToday()
	if err != nil {
		return fmt.Errorf("failed to get connection count: %w", err)
	}

	if todayCount >= c.cfg.Limits.MaxConnectionsPerDay {
		c.logger.Warn("Daily connection limit reached (%d/%d)", todayCount, c.cfg.Limits.MaxConnectionsPerDay)
		return errors.New("daily connection limit reached")
	}

	remaining := c.cfg.Limits.MaxConnectionsPerDay - todayCount
	c.logger.Info("Can send %d more connections today", remaining)

	sent := 0
	for i, profile := range profiles {
		if sent >= remaining {
			c.logger.Info("Reached daily limit")
			break
		}

		// Check if already sent
		alreadySent, err := c.store.IsConnectionSent(profile.URL)
		if err != nil {
			c.logger.Error("Failed to check connection status: %v", err)
			continue
		}

		if alreadySent {
			c.logger.Debug("Already sent connection to %s, skipping", profile.Name)
			continue
		}

		c.logger.Info("[%d/%d] Sending connection to: %s (%s)", i+1, len(profiles), profile.Name, profile.Title)

		// Personalize note
		personalizedNote := c.personalizeNote(note, profile)

		// Send connection request
		if err := c.sendConnection(profile, personalizedNote); err != nil {
			c.logger.Error("Failed to send connection to %s: %v", profile.Name, err)
			continue
		}

		// Save to database
		if err := c.store.SaveConnectionRequest(profile.URL, profile.Name, personalizedNote); err != nil {
			c.logger.Error("Failed to save connection request: %v", err)
		}

		sent++
		c.logger.LogAction("CONNECTION_SENT", map[string]interface{}{
			"name": profile.Name,
			"url":  profile.URL,
		})

		// Random delay between requests
		stealth.HumanDelay(
			c.cfg.Delays.MinActionDelayMs*2,
			c.cfg.Delays.MaxActionDelayMs*2,
		)

		// Occasional break
		stealth.RandomBreak()
	}

	c.logger.Info("Completed: sent %d connection requests", sent)
	return nil
}

func (c *Connector) sendConnection(profile search.Profile, note string) error {
	// Navigate to profile
	c.logger.Debug("Navigating to profile: %s", profile.URL)
	if err := c.page.Navigate(profile.URL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	if err := c.page.WaitLoad(); err != nil {
		return err
	}

	stealth.RandomDelay(2000, 4000)

	// Random scrolling to appear human
	if c.cfg.Stealth.EnableRandomScrolling {
		stealth.PageThroughContent(c.page, 2)
	}

	// Find Connect button
	connectButton, err := c.findConnectButton()
	if err != nil {
		return fmt.Errorf("failed to find connect button: %w", err)
	}

	// Scroll to button
	stealth.ScrollToElement(c.page, connectButton)
	stealth.RandomDelay(500, 1000)

	// Hover before clicking
	if c.cfg.Stealth.EnableMouseHovering {
		stealth.HoverElement(c.page, connectButton)
		stealth.RandomDelay(200, 500)
	}

	// Click Connect
	c.logger.Debug("Clicking Connect button")
	if err := stealth.HumanClick(c.page, connectButton); err != nil {
		return err
	}

	stealth.RandomDelay(1000, 2000)

	// Check if note dialog appeared
	if c.hasNoteDialog() {
		if err := c.addNote(note); err != nil {
			c.logger.Warn("Failed to add note: %v", err)
		}
	}

	// Click Send
	if err := c.clickSend(); err != nil {
		return err
	}

	c.logger.Info("Connection request sent successfully")
	return nil
}

func (c *Connector) findConnectButton() (*rod.Element, error) {
	// Try different selectors
	selectors := []string{
		"button[aria-label*='Connect']",
		"button[aria-label*='Invite']",
		".pvs-profile-actions button:has-text('Connect')",
		"button:has-text('Connect')",
	}

	for _, selector := range selectors {
		if btn, err := c.page.Element(selector); err == nil {
			return btn, nil
		}
	}

	return nil, errors.New("connect button not found")
}

func (c *Connector) hasNoteDialog() bool {
	_, err := c.page.Element("#custom-message")
	return err == nil
}

func (c *Connector) addNote(note string) error {
	// Find Add a note button
	addNoteBtn, err := c.page.Element("button[aria-label='Add a note']")
	if err != nil {
		return err
	}

	stealth.HumanClick(c.page, addNoteBtn)
	stealth.RandomDelay(500, 1000)

	// Find note textarea
	noteField, err := c.page.Element("#custom-message")
	if err != nil {
		return err
	}

	// Type note
	c.logger.Debug("Adding personalized note")
	if err := stealth.HumanType(
		c.page,
		noteField,
		note,
		c.cfg.Delays.MinTypingDelayMs,
		c.cfg.Delays.MaxTypingDelayMs,
		c.cfg.Stealth.TypoProbability,
	); err != nil {
		return err
	}

	stealth.RandomDelay(500, 1000)
	return nil
}

func (c *Connector) clickSend() error {
	sendButton, err := c.page.Element("button[aria-label='Send now']")
	if err != nil {
		sendButton, err = c.page.Element("button[aria-label='Send invitation']")
		if err != nil {
			return errors.New("send button not found")
		}
	}

	stealth.RandomDelay(500, 1000)
	return stealth.HumanClick(c.page, sendButton)
}

func (c *Connector) personalizeNote(template string, profile search.Profile) string {
	note := template
	note = strings.ReplaceAll(note, "{name}", profile.Name)
	note = strings.ReplaceAll(note, "{title}", profile.Title)
	note = strings.ReplaceAll(note, "{location}", profile.Location)

	// Truncate if too long
	if len(note) > c.cfg.Limits.ConnectionNoteMaxLen {
		note = note[:c.cfg.Limits.ConnectionNoteMaxLen-3] + "..."
	}

	return note
}
