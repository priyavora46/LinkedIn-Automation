package stealth

import (
	"math/rand"
	"time"
)

// RandomDelay adds a random delay between min and max milliseconds
func RandomDelay(minMs, maxMs int) {
	delay := time.Duration(minMs+rand.Intn(maxMs-minMs)) * time.Millisecond
	time.Sleep(delay)
}

// HumanDelay simulates human-like delay with occasional longer pauses
func HumanDelay(baseMinMs, baseMaxMs int) {
	delay := baseMinMs + rand.Intn(baseMaxMs-baseMinMs)

	// 10% chance of longer delay (distraction/thinking)
	if rand.Float64() < 0.1 {
		delay += 1000 + rand.Intn(2000)
	}

	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// IsBusinessHours checks if current time is within business hours
func IsBusinessHours(startHour, endHour int) bool {
	now := time.Now()
	hour := now.Hour()

	// Check if weekday
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	return hour >= startHour && hour < endHour
}

// WaitForBusinessHours blocks until business hours
func WaitForBusinessHours(startHour, endHour int) {
	for !IsBusinessHours(startHour, endHour) {
		now := time.Now()

		// Calculate time until next business hour
		nextStart := time.Date(now.Year(), now.Month(), now.Day(), startHour, 0, 0, 0, now.Location())

		// If past business hours today, wait until tomorrow
		if now.Hour() >= endHour {
			nextStart = nextStart.Add(24 * time.Hour)
		}

		// If weekend, wait until Monday
		if now.Weekday() == time.Saturday {
			nextStart = nextStart.Add(48 * time.Hour)
		} else if now.Weekday() == time.Sunday {
			nextStart = nextStart.Add(24 * time.Hour)
		}

		waitDuration := time.Until(nextStart)
		time.Sleep(waitDuration)
	}
}

// RandomBreak simulates taking a random break
func RandomBreak() {
	// 5% chance of taking a break
	if rand.Float64() < 0.05 {
		breakDuration := time.Duration(2+rand.Intn(5)) * time.Minute
		time.Sleep(breakDuration)
	}
}

// ThrottleAction ensures minimum time between actions
func ThrottleAction(lastActionTime time.Time, minInterval time.Duration) {
	elapsed := time.Since(lastActionTime)
	if elapsed < minInterval {
		time.Sleep(minInterval - elapsed)
	}
}

// ExponentialBackoff implements retry delay with exponential increase
func ExponentialBackoff(attempt int, baseDelay time.Duration, maxDelay time.Duration) time.Duration {
	delay := baseDelay * time.Duration(1<<uint(attempt))
	if delay > maxDelay {
		delay = maxDelay
	}

	// Add jitter
	jitter := time.Duration(rand.Int63n(int64(delay / 4)))
	return delay + jitter
}

// RandomizeSchedule returns a random time within a window
func RandomizeSchedule(baseTime time.Time, windowMinutes int) time.Time {
	offset := rand.Intn(windowMinutes * 60)
	return baseTime.Add(time.Duration(offset) * time.Second)
}
