package stealth

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

// HumanType simulates human typing with realistic patterns
func HumanType(page *rod.Page, el *rod.Element, text string, minDelay, maxDelay int, typoProb float64) error {
	// Focus the element first
	if err := el.Focus(); err != nil {
		return err
	}

	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	for i, char := range text {
		// Occasionally make a typo
		if typoProb > 0 && rand.Float64() < typoProb && i < len(text)-1 {
			// Type wrong character
			wrongChar := rune('a' + rand.Intn(26))
			page.Keyboard.Type(input.Key(wrongChar))
			time.Sleep(time.Duration(minDelay+rand.Intn(maxDelay-minDelay)) * time.Millisecond)

			// Pause (realize mistake)
			time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)

			// Backspace
			page.Keyboard.Press(input.Backspace)
			time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		}

		// Type the correct character
		page.Keyboard.Type(input.Key(char))

		// Variable typing speed
		var delay time.Duration
		if char == ' ' || char == '.' || char == ',' {
			// Longer pause after punctuation
			delay = time.Duration(minDelay+rand.Intn(maxDelay-minDelay)+50) * time.Millisecond
		} else {
			delay = time.Duration(minDelay+rand.Intn(maxDelay-minDelay)) * time.Millisecond
		}

		// Occasionally longer pauses (thinking)
		if rand.Float64() < 0.05 {
			delay += time.Duration(300+rand.Intn(500)) * time.Millisecond
		}

		time.Sleep(delay)
	}

	// Brief pause after typing
	time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)

	return nil
}

// TypeWithBackspace simulates typing with occasional backspacing
func TypeWithBackspace(page *rod.Page, el *rod.Element, text string) error {
	el.Focus()
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	words := splitIntoWords(text)

	for i, word := range words {
		// Type word
		for _, char := range word {
			page.Keyboard.Type(input.Key(char))
			time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)
		}

		// Sometimes backspace and retype
		if rand.Float64() < 0.1 && len(word) > 3 {
			backspaceCount := 1 + rand.Intn(3)
			for j := 0; j < backspaceCount; j++ {
				page.Keyboard.Press(input.Backspace)
				time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
			}

			// Retype
			retyped := word[len(word)-backspaceCount:]
			for _, char := range retyped {
				page.Keyboard.Type(input.Key(char))
				time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)
			}
		}

		// Add space between words
		if i < len(words)-1 {
			page.Keyboard.Type(input.Key(' '))
			time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
		}
	}

	return nil
}

func splitIntoWords(text string) []string {
	var words []string
	var current string

	for _, char := range text {
		if char == ' ' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}

// SimulateThinking adds a pause to simulate thinking before typing
func SimulateThinking() {
	thinkTime := time.Duration(500+rand.Intn(1500)) * time.Millisecond
	time.Sleep(thinkTime)
}
