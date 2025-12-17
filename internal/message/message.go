package message

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-rod/rod"

	"linkedin-automation/config"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/stealth"
	"linkedin-automation/internal/storage"
)

type Messenger struct {
	page   *rod.Page
	cfg    *config.Config
	logger *logger.Logger
	store  *storage.Store
}

func New(page *rod.Page, cfg *config.Config, log *logger.Logger, store *storage.Store) *Messenger {
	return &Messenger{
		page:   page,
		cfg:    cfg,
		logger: log,
		store:  store,
	}
}

func (m *Messenger) SendFollowUpMessages(messageTemplate string) error {
	m.logger.Info("Checking for accepted connections")

	// Get pending connections
	connections, err := m.store.GetPendingConnections()
	if err != nil {
		return fmt.Errorf("failed to get pending connections: %w", err)
	}

	m.logger.Info("Found %d pending connections", len(connections))

	// Check message limit
	todayCount, err := m.store.GetMessagesCountToday()
	if err != nil {
		return err
	}

	if todayCount >= m.cfg.Limits.MaxMessagesPerDay {
		m.logger.Warn("Daily message limit reached")
		return errors.New("daily message limit reached")
	}

	sent := 0
	remaining := m.cfg.Limits.MaxMessagesPerDay - todayCount

	for _, conn := range connections {
		if sent >= remaining {
			break
		}

		// Check if connection is accepted
		accepted, err := m.checkConnectionAccepted(conn.ProfileURL)
		if err != nil {
			m.logger.Error("Failed to check connection status: %v", err)
			continue
		}

		if !accepted {
			continue
		}

		m.logger.Info("Connection accepted: %s. Sending follow-up message", conn.Name)

		// Send message
		if err := m.sendMessage(conn.ProfileURL, conn.Name, messageTemplate); err != nil {
			m.logger.Error("Failed to send message: %v", err)
			continue
		}

		// Update database
		m.store.MarkConnectionAccepted(conn.ProfileURL)
		m.store.SaveMessage(conn.ProfileURL, messageTemplate)

		sent++
		m.logger.LogAction("MESSAGE_SENT", map[string]interface{}{
			"name": conn.Name,
			"url":  conn.ProfileURL,
		})

		// Delay between messages
		stealth.HumanDelay(
			m.cfg.Delays.MinActionDelayMs*3,
			m.cfg.Delays.MaxActionDelayMs*3,
		)
	}

	m.logger.Info("Sent %d follow-up messages", sent)
	return nil
}

func (m *Messenger) checkConnectionAccepted(profileURL string) (bool, error) {
	// Navigate to profile
	if err := m.page.Navigate(profileURL); err != nil {
		return false, err
	}

	m.page.WaitLoad()
	stealth.RandomDelay(1000, 2000)

	// Check if "Message" button exists (indicates connected)
	_, err := m.page.Element("button[aria-label*='Message']")
	return err == nil, nil
}

func (m *Messenger) sendMessage(profileURL, name, template string) error {
	// Navigate to messaging
	messagingURL := fmt.Sprintf("https://www.linkedin.com/messaging/thread/new/?recipient=%s", extractProfileID(profileURL))

	if err := m.page.Navigate(messagingURL); err != nil {
		return err
	}

	m.page.WaitLoad()
	stealth.RandomDelay(2000, 3000)

	// Find message compose box
	composeBox, err := m.findComposeBox()
	if err != nil {
		return err
	}

	// Click to focus
	stealth.HumanClick(m.page, composeBox)
	stealth.SimulateThinking()

	// Type message
	m.logger.Debug("Typing message")
	if err := stealth.HumanType(
		m.page,
		composeBox,
		template,
		m.cfg.Delays.MinTypingDelayMs,
		m.cfg.Delays.MaxTypingDelayMs,
		m.cfg.Stealth.TypoProbability,
	); err != nil {
		return err
	}

	stealth.RandomDelay(1000, 2000)

	// Find and click send button
	sendBtn, err := m.page.Element("button[type='submit']")
	if err != nil {
		return errors.New("send button not found")
	}

	return stealth.HumanClick(m.page, sendBtn)
}

func (m *Messenger) findComposeBox() (*rod.Element, error) {
	selectors := []string{
		".msg-form__contenteditable",
		"div[role='textbox']",
		".msg-form__msg-content-container",
	}

	for _, sel := range selectors {
		if el, err := m.page.Element(sel); err == nil {
			return el, nil
		}
	}

	return nil, errors.New("compose box not found")
}

func extractProfileID(profileURL string) string {
	// Extract profile ID from URL
	// Example: https://www.linkedin.com/in/john-doe-123456/ -> john-doe-123456
	parts := strings.Split(profileURL, "/in/")
	if len(parts) < 2 {
		return ""
	}

	id := parts[1]
	id = strings.TrimSuffix(id, "/")
	return id
}
