package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"linkedin-automation/config"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/stealth"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Authenticator struct {
	page   *rod.Page
	cfg    *config.Config
	logger *logger.Logger
}

func New(page *rod.Page, cfg *config.Config, log *logger.Logger) *Authenticator {
	return &Authenticator{
		page:   page,
		cfg:    cfg,
		logger: log,
	}
}

func (a *Authenticator) Login() error {
	a.logger.Info("Starting login process")

	// Attempt session restore
	if err := a.loadSession(); err == nil {
		a.logger.Info("Loaded saved session")
		if a.isLoggedIn() {
			a.logger.Info("Session is still valid")
			return nil
		}
	}

	// Navigate to LinkedIn login page
	a.logger.Info("Navigating to login page")
	if err := a.page.Navigate(a.cfg.LinkedIn.LoginURL); err != nil {
		return fmt.Errorf("failed to navigate to login: %w", err)
	}

	if err := a.page.WaitLoad(); err != nil {
		return err
	}

	// 🔥 FORCE DESKTOP VIEWPORT (Rod New API)
	if err := a.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:             1366,
		Height:            768,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}); err != nil {
		a.logger.Warn("Failed to set viewport: %v", err)
	}

	// 🔥 Reset zoom (Windows Chromium fix)
	_, _ = a.page.Eval(`document.body.style.zoom = "100%"`)

	stealth.RandomDelay(1000, 2000)

	// Locate email field
	emailField, err := a.page.Element("#username")
	if err != nil {
		return fmt.Errorf("failed to find email field: %w", err)
	}

	a.logger.Debug("Clicking email field")
	if err := stealth.HumanClick(a.page, emailField); err != nil {
		return err
	}

	stealth.SimulateThinking()

	a.logger.Debug("Typing email")
	if err := stealth.HumanType(
		a.page,
		emailField,
		a.cfg.Creds.Email,
		a.cfg.Delays.MinTypingDelayMs,
		a.cfg.Delays.MaxTypingDelayMs,
		0,
	); err != nil {
		return err
	}

	stealth.RandomDelay(500, 1000)

	// Locate password field
	passwordField, err := a.page.Element("#password")
	if err != nil {
		return fmt.Errorf("failed to find password field: %w", err)
	}

	a.logger.Debug("Clicking password field")
	if err := stealth.HumanClick(a.page, passwordField); err != nil {
		return err
	}

	stealth.SimulateThinking()

	a.logger.Debug("Typing password")
	if err := stealth.HumanType(
		a.page,
		passwordField,
		a.cfg.Creds.Password,
		a.cfg.Delays.MinTypingDelayMs,
		a.cfg.Delays.MaxTypingDelayMs,
		0,
	); err != nil {
		return err
	}

	stealth.RandomDelay(1000, 2000)

	// Submit login form
	loginButton, err := a.page.Element("button[type='submit']")
	if err != nil {
		return fmt.Errorf("failed to find login button: %w", err)
	}

	a.logger.Debug("Clicking login button")
	if err := stealth.HumanClick(a.page, loginButton); err != nil {
		return err
	}

	a.logger.Info("Waiting for login to complete")
	time.Sleep(5 * time.Second)

	// Detect CAPTCHA / 2FA
	if a.hasSecurityChallenge() {
		a.logger.Warn("Security challenge detected")
		return errors.New("security challenge detected – manual intervention required")
	}

	if !a.isLoggedIn() {
		a.logger.Error("Login failed")
		return errors.New("login failed – check credentials")
	}

	a.logger.Info("Login successful")

	if err := a.saveSession(); err != nil {
		a.logger.Warn("Failed to save session: %v", err)
	}

	return nil
}

func (a *Authenticator) isLoggedIn() bool {
	url := a.page.MustInfo().URL
	return contains(url, "/feed") ||
		contains(url, "/mynetwork") ||
		contains(url, "/messaging")
}

func (a *Authenticator) hasSecurityChallenge() bool {
	selectors := []string{
		"#input__phone_verification_pin",
		"#captcha",
		".challenge-dialog",
		"input[name='pin']",
	}

	for _, s := range selectors {
		if _, err := a.page.Element(s); err == nil {
			return true
		}
	}
	return false
}

func (a *Authenticator) saveSession() error {
	cookies := a.page.MustCookies()
	var stored []*proto.NetworkCookieParam

	for _, c := range cookies {
		stored = append(stored, &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
			SameSite: c.SameSite,
			Expires:  c.Expires,
		})
	}

	data, err := json.Marshal(stored)
	if err != nil {
		return err
	}

	os.MkdirAll("data", 0755)
	return os.WriteFile(a.cfg.Storage.SessionCookiePath, data, 0600)
}

func (a *Authenticator) loadSession() error {
	data, err := os.ReadFile(a.cfg.Storage.SessionCookiePath)
	if err != nil {
		return err
	}

	var cookies []*proto.NetworkCookieParam
	if err := json.Unmarshal(data, &cookies); err != nil {
		return err
	}

	if err := a.page.SetCookies(cookies); err != nil {
		return err
	}

	a.page.Navigate(a.cfg.LinkedIn.BaseURL)
	a.page.WaitLoad()
	return nil
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
