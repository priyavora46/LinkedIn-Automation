package stealth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
)

// HumanScroll simulates human-like scrolling behavior
func HumanScroll(page *rod.Page, direction string, distance int) error {
	steps := 5 + rand.Intn(10)
	stepDistance := float64(distance) / float64(steps)

	for i := 0; i < steps; i++ {
		// Variable speed - acceleration and deceleration
		var speed float64
		progress := float64(i) / float64(steps)

		if progress < 0.3 {
			// Accelerating
			speed = progress * 3
		} else if progress > 0.7 {
			// Decelerating
			speed = (1 - progress) * 3
		} else {
			// Constant speed
			speed = 1.0
		}

		currentStep := stepDistance * (0.5 + speed)

		if direction == "down" {
			page.MustEval(fmt.Sprintf(`window.scrollBy(0, %f)`, currentStep))
		} else {
			page.MustEval(fmt.Sprintf(`window.scrollBy(0, %f)`, -currentStep))
		}

		// Variable delay between scroll steps
		delay := time.Duration(30+rand.Intn(50)) * time.Millisecond
		time.Sleep(delay)
	}

	// Occasional scroll back
	if rand.Float64() < 0.15 {
		time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)
		smallScrollBack := float64(distance) * 0.1
		if direction == "down" {
			page.MustEval(fmt.Sprintf(`window.scrollBy(0, %f)`, -smallScrollBack))
		} else {
			page.MustEval(fmt.Sprintf(`window.scrollBy(0, %f)`, smallScrollBack))
		}
	}

	return nil
}

// ScrollToElement scrolls to make an element visible
func ScrollToElement(page *rod.Page, el *rod.Element) error {
	// Use Rod's built-in scroll into view
	return el.ScrollIntoView()
}

// RandomScroll performs random scrolling to appear human-like
func RandomScroll(page *rod.Page) error {
	// Random scroll distance
	distance := 100 + rand.Intn(300)

	// Random direction (70% down, 30% up)
	direction := "down"
	if rand.Float64() < 0.3 {
		direction = "up"
	}

	return HumanScroll(page, direction, distance)
}

// ScrollToBottom scrolls to the bottom of the page naturally
func ScrollToBottom(page *rod.Page) error {
	// Get page height
	totalHeight := page.MustEval(`() => document.body.scrollHeight`).Int()
	currentScroll := 0

	for currentScroll < totalHeight {
		scrollAmount := 200 + rand.Intn(300)

		if err := HumanScroll(page, "down", scrollAmount); err != nil {
			return err
		}

		currentScroll += scrollAmount

		// Random pause while "reading"
		time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)

		// Occasionally scroll back up slightly
		if rand.Float64() < 0.2 {
			HumanScroll(page, "up", 50+rand.Intn(100))
			time.Sleep(time.Duration(300+rand.Intn(500)) * time.Millisecond)
		}
	}

	return nil
}

// PageThroughContent simulates reading through page content
func PageThroughContent(page *rod.Page, sections int) error {
	for i := 0; i < sections; i++ {
		// Scroll down
		if err := HumanScroll(page, "down", 300+rand.Intn(400)); err != nil {
			return err
		}

		// Pause to "read"
		readTime := time.Duration(1000+rand.Intn(3000)) * time.Millisecond
		time.Sleep(readTime)

		// Occasionally scroll back to reread
		if rand.Float64() < 0.25 {
			HumanScroll(page, "up", 100+rand.Intn(200))
			time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)
			HumanScroll(page, "down", 100+rand.Intn(200))
		}
	}

	return nil
}
