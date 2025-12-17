package search

import (
	"fmt"
	"linkedin-automation/config"
	"linkedin-automation/internal/logger"
	"linkedin-automation/internal/stealth"
	"net/url"
	"strings"

	"github.com/go-rod/rod"
)

type Searcher struct {
	page   *rod.Page
	cfg    *config.Config
	logger *logger.Logger
}

type Profile struct {
	URL      string
	Name     string
	Title    string
	Location string
}

func New(page *rod.Page, cfg *config.Config, log *logger.Logger) *Searcher {
	return &Searcher{
		page:   page,
		cfg:    cfg,
		logger: log,
	}
}

func (s *Searcher) SearchPeople(query, location, company string, maxResults int) ([]Profile, error) {
	s.logger.Info("Starting people search: query=%s, location=%s, company=%s", query, location, company)

	// Build search URL
	searchURL := s.buildSearchURL(query, location, company)
	s.logger.Debug("Search URL: %s", searchURL)

	// Navigate to search
	if err := s.page.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search: %w", err)
	}

	if err := s.page.WaitLoad(); err != nil {
		return nil, err
	}

	stealth.RandomDelay(2000, 4000)

	// Scroll to load results
	if s.cfg.Stealth.EnableRandomScrolling {
		stealth.HumanScroll(s.page, "down", 300)
		stealth.RandomDelay(1000, 2000)
	}

	var profiles []Profile
	seenURLs := make(map[string]bool)
	page := 1

	for len(profiles) < maxResults {
		s.logger.Info("Processing page %d (collected %d/%d profiles)", page, len(profiles), maxResults)

		// Extract profiles from current page
		pageProfiles, err := s.extractProfiles()
		if err != nil {
			s.logger.Error("Failed to extract profiles: %v", err)
			break
		}

		// Filter duplicates
		for _, p := range pageProfiles {
			if !seenURLs[p.URL] && len(profiles) < maxResults {
				profiles = append(profiles, p)
				seenURLs[p.URL] = true
			}
		}

		// Try to go to next page
		if len(profiles) < maxResults {
			if !s.hasNextPage() {
				s.logger.Info("No more pages available")
				break
			}

			if err := s.goToNextPage(); err != nil {
				s.logger.Warn("Failed to go to next page: %v", err)
				break
			}

			page++
		} else {
			break
		}
	}

	s.logger.Info("Search completed: found %d profiles", len(profiles))
	return profiles, nil
}

func (s *Searcher) buildSearchURL(query, location, company string) string {
	baseURL := s.cfg.LinkedIn.SearchURL
	params := url.Values{}

	// Build keywords
	keywords := []string{}
	if query != "" {
		keywords = append(keywords, query)
	}
	if company != "" {
		keywords = append(keywords, fmt.Sprintf("company:\"%s\"", company))
	}

	if len(keywords) > 0 {
		params.Add("keywords", strings.Join(keywords, " "))
	}

	if location != "" {
		params.Add("location", location)
	}

	params.Add("origin", "FACETED_SEARCH")

	return baseURL + "?" + params.Encode()
}

func (s *Searcher) extractProfiles() ([]Profile, error) {
	// Wait for search results to load
	stealth.RandomDelay(1000, 2000)

	// Find all profile cards
	elements, err := s.page.Elements(".reusable-search__result-container")
	if err != nil {
		return nil, fmt.Errorf("failed to find profile elements: %w", err)
	}

	var profiles []Profile

	for _, el := range elements {
		profile := Profile{}

		// Extract profile URL
		linkEl, err := el.Element("a.app-aware-link")
		if err != nil {
			continue
		}

		href, err := linkEl.Property("href")
		if err != nil {
			continue
		}
		profile.URL = href.String()

		// Extract name
		nameEl, err := el.Element(".entity-result__title-text a span[aria-hidden='true']")
		if err == nil {
			name, _ := nameEl.Text()
			profile.Name = strings.TrimSpace(name)
		}

		// Extract title
		titleEl, err := el.Element(".entity-result__primary-subtitle")
		if err == nil {
			title, _ := titleEl.Text()
			profile.Title = strings.TrimSpace(title)
		}

		// Extract location
		locationEl, err := el.Element(".entity-result__secondary-subtitle")
		if err == nil {
			location, _ := locationEl.Text()
			profile.Location = strings.TrimSpace(location)
		}

		if profile.URL != "" {
			profiles = append(profiles, profile)
		}
	}

	s.logger.Debug("Extracted %d profiles from page", len(profiles))
	return profiles, nil
}

func (s *Searcher) hasNextPage() bool {
	nextButton, err := s.page.Element("button[aria-label='Next']")
	if err != nil {
		return false
	}

	disabled, _ := nextButton.Property("disabled")
	return disabled.Nil()
}

func (s *Searcher) goToNextPage() error {
	s.logger.Debug("Going to next page")

	nextButton, err := s.page.Element("button[aria-label='Next']")
	if err != nil {
		return err
	}

	// Scroll to button
	stealth.ScrollToElement(s.page, nextButton)
	stealth.RandomDelay(500, 1000)

	// Click next
	if err := stealth.HumanClick(s.page, nextButton); err != nil {
		return err
	}

	// Wait for page to load
	stealth.RandomDelay(2000, 4000)

	// Random scrolling on new page
	if s.cfg.Stealth.EnableRandomScrolling {
		stealth.RandomScroll(s.page)
	}

	return nil
}
