package browser

import (
	"fmt"
	"linkedin-automation/config"
	"linkedin-automation/internal/logger"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Browser struct {
	page   *rod.Page
	cfg    *config.Config
	logger *logger.Logger
}

func New(cfg *config.Config, log *logger.Logger) (*Browser, error) {
	// Launch browser
	u := launcher.New().
		Headless(cfg.Browser.Headless).
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()

	// Create page
	page := browser.MustPage("")

	// Set viewport
	page.MustSetViewport(cfg.Browser.Width, cfg.Browser.Height, 1, false)

	// Apply stealth techniques
	if err := applyStealth(page, cfg); err != nil {
		return nil, fmt.Errorf("failed to apply stealth: %w", err)
	}

	log.Info("Browser initialized successfully")

	return &Browser{
		page:   page,
		cfg:    cfg,
		logger: log,
	}, nil
}

func applyStealth(page *rod.Page, cfg *config.Config) error {
	// Override navigator.webdriver
	page.MustEval(`() => {
		Object.defineProperty(navigator, 'webdriver', {
			get: () => false
		});
	}`)

	// Set user agent
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: cfg.Browser.UserAgent,
	})

	// Override plugins
	page.MustEval(`() => {
		Object.defineProperty(navigator, 'plugins', {
			get: () => [1, 2, 3, 4, 5]
		});
	}`)

	// Override languages
	page.MustEval(`() => {
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en']
		});
	}`)

	// Override permissions
	page.MustEval(`() => {
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);
	}`)

	// Chrome detection
	page.MustEval(`() => {
		window.chrome = {
			runtime: {}
		};
	}`)

	return nil
}

func (b *Browser) Page() *rod.Page {
	return b.page
}

func (b *Browser) Navigate(url string) error {
	b.logger.Debug("Navigating to: %s", url)
	return b.page.Navigate(url)
}

func (b *Browser) WaitLoad() error {
	return b.page.WaitLoad()
}

func (b *Browser) Close() error {
	b.logger.Info("Closing browser")
	if b.page != nil {
		return b.page.Close()
	}
	return nil
}

func (b *Browser) Screenshot(path string) error {
	data, err := b.page.Screenshot(false, nil)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
