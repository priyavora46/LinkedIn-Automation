package stealth

import (
	"math/rand"
	"time"
)

// ActivityScheduler manages realistic activity timing patterns
type ActivityScheduler struct {
	timezone         *time.Location
	workHoursStart   int // 9 AM
	workHoursEnd     int // 6 PM
	breakTimes       []TimeWindow
	lastActivityTime time.Time
	dailyActionCount int
	rand             *rand.Rand
}

// TimeWindow represents a time range
type TimeWindow struct {
	Start time.Time
	End   time.Time
}

// NewActivityScheduler creates a new scheduler with realistic timing
func NewActivityScheduler(timezone string) (*ActivityScheduler, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	return &ActivityScheduler{
		timezone:       loc,
		workHoursStart: 9,
		workHoursEnd:   18,
		breakTimes: []TimeWindow{
			// Lunch break: 12-1 PM with slight variation
			{},
		},
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// IsWorkingHours checks if current time is within working hours
func (as *ActivityScheduler) IsWorkingHours() bool {
	now := time.Now().In(as.timezone)
	hour := now.Hour()

	// Weekend check
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		// Occasionally active on weekends (10% chance)
		return as.rand.Float64() < 0.10
	}

	// Check if within working hours with some flexibility
	startHour := as.workHoursStart + as.rand.Intn(2) - 1 // 8-10 AM start
	endHour := as.workHoursEnd + as.rand.Intn(3) - 1     // 5-8 PM end

	return hour >= startHour && hour < endHour
}

// IsBreakTime checks if it's a typical break period
func (as *ActivityScheduler) IsBreakTime() bool {
	now := time.Now().In(as.timezone)
	hour := now.Hour()
	minute := now.Minute()

	// Lunch break (12:00 - 13:30) with variation
	if hour == 12 || (hour == 13 && minute < 30) {
		return as.rand.Float64() < 0.70 // 70% chance to be on break
	}

	// Morning break (10:00 - 10:15)
	if hour == 10 && minute < 15 {
		return as.rand.Float64() < 0.40 // 40% chance
	}

	// Afternoon break (15:00 - 15:15)
	if hour == 15 && minute < 15 {
		return as.rand.Float64() < 0.40 // 40% chance
	}

	return false
}

// ShouldTakeBreak determines if a break should be taken based on activity
func (as *ActivityScheduler) ShouldTakeBreak() bool {
	now := time.Now()

	// If last activity was recent, occasionally take a break
	if !as.lastActivityTime.IsZero() {
		duration := now.Sub(as.lastActivityTime)

		// After 45-90 minutes of activity, higher chance of break
		if duration.Minutes() >= 45 {
			chance := 0.15 + (duration.Minutes()-45)/180.0 // Increases over time
			if as.rand.Float64() < chance {
				return true
			}
		}
	}

	// Random micro-breaks (5% chance)
	return as.rand.Float64() < 0.05
}

// GetBreakDuration returns a realistic break duration
func (as *ActivityScheduler) GetBreakDuration() time.Duration {
	breakType := as.rand.Float64()

	switch {
	case breakType < 0.60: // 60% - Quick break (2-5 minutes)
		return time.Duration(2+as.rand.Intn(4)) * time.Minute

	case breakType < 0.85: // 25% - Medium break (5-15 minutes)
		return time.Duration(5+as.rand.Intn(11)) * time.Minute

	default: // 15% - Long break (15-45 minutes)
		return time.Duration(15+as.rand.Intn(31)) * time.Minute
	}
}

// GetThinkTime returns realistic delay between actions (human cognitive processing)
func (as *ActivityScheduler) GetThinkTime() time.Duration {
	// Base think time: 1-5 seconds
	base := 1000 + as.rand.Intn(4000)

	// Add occasional longer pauses (15% chance)
	if as.rand.Float64() < 0.15 {
		base += 3000 + as.rand.Intn(7000) // Additional 3-10 seconds
	}

	return time.Duration(base) * time.Millisecond
}

// GetActionInterval returns time between major actions (profile views, messages)
func (as *ActivityScheduler) GetActionInterval() time.Duration {
	// Realistic interval: 15-90 seconds between actions
	baseSeconds := 15 + as.rand.Intn(76)

	// Vary based on time of day
	now := time.Now().In(as.timezone)
	hour := now.Hour()

	// Slower during early morning and late evening
	if hour < 10 || hour > 17 {
		baseSeconds += as.rand.Intn(30) // Add 0-30 seconds
	}

	// Faster mid-day (peak productivity)
	if hour >= 10 && hour <= 15 {
		baseSeconds -= as.rand.Intn(10) // Subtract 0-10 seconds
		if baseSeconds < 10 {
			baseSeconds = 10
		}
	}

	return time.Duration(baseSeconds) * time.Second
}

// GetScrollDelay returns realistic scroll timing
func (as *ActivityScheduler) GetScrollDelay() time.Duration {
	// Quick scroll: 200-800ms
	if as.rand.Float64() < 0.70 {
		return time.Duration(200+as.rand.Intn(600)) * time.Millisecond
	}
	// Slower scroll/reading: 1-3 seconds
	return time.Duration(1000+as.rand.Intn(2000)) * time.Millisecond
}

// GetTypingDelay returns realistic keystroke interval
func (as *ActivityScheduler) GetTypingDelay() time.Duration {
	// Average typing: 50-200ms per character
	base := 50 + as.rand.Intn(151)

	// Occasional longer pauses (thinking while typing)
	if as.rand.Float64() < 0.10 {
		base += 300 + as.rand.Intn(700) // 300-1000ms pause
	}

	return time.Duration(base) * time.Millisecond
}

// GetPageLoadWait returns realistic page load waiting time
func (as *ActivityScheduler) GetPageLoadWait() time.Duration {
	// Wait for page to fully load and render: 2-5 seconds
	return time.Duration(2000+as.rand.Intn(3000)) * time.Millisecond
}

// RecordActivity updates the last activity timestamp
func (as *ActivityScheduler) RecordActivity() {
	as.lastActivityTime = time.Now()
	as.dailyActionCount++
}

// ResetDailyCount resets the daily action counter (call at midnight)
func (as *ActivityScheduler) ResetDailyCount() {
	as.dailyActionCount = 0
}

// GetDailyActionCount returns current action count for the day
func (as *ActivityScheduler) GetDailyActionCount() int {
	return as.dailyActionCount
}

// SimulateHumanRhythm adds natural rhythm variations
func (as *ActivityScheduler) SimulateHumanRhythm() time.Duration {
	hour := time.Now().In(as.timezone).Hour()

	// Morning: slower (just starting)
	if hour >= 9 && hour < 11 {
		return time.Duration(2000+as.rand.Intn(3000)) * time.Millisecond
	}

	// Mid-day: faster (peak productivity)
	if hour >= 11 && hour < 14 {
		return time.Duration(800+as.rand.Intn(1200)) * time.Millisecond
	}

	// Afternoon: moderate
	if hour >= 14 && hour < 16 {
		return time.Duration(1500+as.rand.Intn(2000)) * time.Millisecond
	}

	// Late afternoon: slower (fatigue)
	if hour >= 16 && hour < 18 {
		return time.Duration(2500+as.rand.Intn(3500)) * time.Millisecond
	}

	// Default
	return time.Duration(1500+as.rand.Intn(2000)) * time.Millisecond
}

// GetRandomDelay returns a general-purpose random delay
func (as *ActivityScheduler) GetRandomDelay(minMs, maxMs int) time.Duration {
	if maxMs <= minMs {
		maxMs = minMs + 1000
	}
	return time.Duration(minMs+as.rand.Intn(maxMs-minMs)) * time.Millisecond
}
