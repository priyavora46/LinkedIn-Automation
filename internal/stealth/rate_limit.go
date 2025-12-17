package stealth

import (
	"fmt"
	"sync"
	"time"
)

// ActionType represents different types of actions
type ActionType string

const (
	ActionProfileView   ActionType = "profile_view"
	ActionConnectionReq ActionType = "connection_request"
	ActionMessage       ActionType = "message"
	ActionSearch        ActionType = "search"
	ActionScroll        ActionType = "scroll"
	ActionLike          ActionType = "like"
	ActionComment       ActionType = "comment"
	ActionPageView      ActionType = "page_view"
)

// RateLimiter manages action quotas and cooldowns
type RateLimiter struct {
	mu                 sync.RWMutex
	limits             map[ActionType]*ActionLimit
	actionHistory      map[ActionType][]time.Time
	dailyResetTime     time.Time
	hourlyResetTime    time.Time
	cooldownUntil      time.Time
	consecutiveActions int
}

// ActionLimit defines limits for a specific action
type ActionLimit struct {
	HourlyMax        int
	DailyMax         int
	MinInterval      time.Duration // Minimum time between same actions
	CooldownAfter    int           // Trigger cooldown after N consecutive actions
	CooldownDuration time.Duration
}

// NewRateLimiter creates a new rate limiter with realistic LinkedIn limits
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits:        make(map[ActionType]*ActionLimit),
		actionHistory: make(map[ActionType][]time.Time),
	}

	// Configure realistic LinkedIn limits based on anti-detection needs
	rl.limits[ActionProfileView] = &ActionLimit{
		HourlyMax:        40,  // Conservative: ~80-100 safe, but 40 is safer
		DailyMax:         200, // Conservative daily limit
		MinInterval:      15 * time.Second,
		CooldownAfter:    10,
		CooldownDuration: 5 * time.Minute,
	}

	rl.limits[ActionConnectionReq] = &ActionLimit{
		HourlyMax:        10, // Very conservative for connection requests
		DailyMax:         50, // LinkedIn typically allows 100-200, but stay safe
		MinInterval:      90 * time.Second,
		CooldownAfter:    5,
		CooldownDuration: 10 * time.Minute,
	}

	rl.limits[ActionMessage] = &ActionLimit{
		HourlyMax:        15,
		DailyMax:         80,
		MinInterval:      60 * time.Second,
		CooldownAfter:    7,
		CooldownDuration: 8 * time.Minute,
	}

	rl.limits[ActionSearch] = &ActionLimit{
		HourlyMax:        30,
		DailyMax:         150,
		MinInterval:      20 * time.Second,
		CooldownAfter:    8,
		CooldownDuration: 3 * time.Minute,
	}

	rl.limits[ActionScroll] = &ActionLimit{
		HourlyMax:        200, // Higher for natural browsing
		DailyMax:         1000,
		MinInterval:      2 * time.Second,
		CooldownAfter:    20,
		CooldownDuration: 2 * time.Minute,
	}

	rl.limits[ActionLike] = &ActionLimit{
		HourlyMax:        25,
		DailyMax:         120,
		MinInterval:      30 * time.Second,
		CooldownAfter:    8,
		CooldownDuration: 5 * time.Minute,
	}

	rl.limits[ActionComment] = &ActionLimit{
		HourlyMax:        8,
		DailyMax:         30,
		MinInterval:      120 * time.Second,
		CooldownAfter:    3,
		CooldownDuration: 15 * time.Minute,
	}

	rl.limits[ActionPageView] = &ActionLimit{
		HourlyMax:        100,
		DailyMax:         500,
		MinInterval:      5 * time.Second,
		CooldownAfter:    15,
		CooldownDuration: 3 * time.Minute,
	}

	rl.resetTimers()
	return rl
}

// CanPerformAction checks if an action is allowed
func (rl *RateLimiter) CanPerformAction(actionType ActionType) (bool, string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if in cooldown period
	if time.Now().Before(rl.cooldownUntil) {
		remaining := time.Until(rl.cooldownUntil)
		return false, fmt.Sprintf("In cooldown period. Wait %v", remaining.Round(time.Second))
	}

	limit, exists := rl.limits[actionType]
	if !exists {
		return true, "" // No limit defined, allow action
	}

	// Clean old history entries
	rl.cleanHistory(actionType)

	history := rl.actionHistory[actionType]
	now := time.Now()

	// Check hourly limit
	hourlyCount := rl.countActionsInWindow(history, time.Hour)
	if hourlyCount >= limit.HourlyMax {
		return false, fmt.Sprintf("Hourly limit reached (%d/%d)", hourlyCount, limit.HourlyMax)
	}

	// Check daily limit
	dailyCount := rl.countActionsInWindow(history, 24*time.Hour)
	if dailyCount >= limit.DailyMax {
		return false, fmt.Sprintf("Daily limit reached (%d/%d)", dailyCount, limit.DailyMax)
	}

	// Check minimum interval
	if len(history) > 0 {
		lastAction := history[len(history)-1]
		if now.Sub(lastAction) < limit.MinInterval {
			remaining := limit.MinInterval - now.Sub(lastAction)
			return false, fmt.Sprintf("Too soon. Wait %v", remaining.Round(time.Second))
		}
	}

	return true, ""
}

// RecordAction records an action and updates counters
func (rl *RateLimiter) RecordAction(actionType ActionType) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[actionType]
	if !exists {
		return fmt.Errorf("no limit defined for action type: %s", actionType)
	}

	now := time.Now()

	// Add to history
	if rl.actionHistory[actionType] == nil {
		rl.actionHistory[actionType] = []time.Time{}
	}
	rl.actionHistory[actionType] = append(rl.actionHistory[actionType], now)

	// Check for consecutive actions triggering cooldown
	rl.consecutiveActions++
	if rl.consecutiveActions >= limit.CooldownAfter {
		rl.cooldownUntil = now.Add(limit.CooldownDuration)
		rl.consecutiveActions = 0
	}

	// Reset timers if needed
	if now.After(rl.hourlyResetTime) {
		rl.hourlyResetTime = now.Add(time.Hour)
	}
	if now.After(rl.dailyResetTime) {
		rl.dailyResetTime = now.Add(24 * time.Hour)
	}

	return nil
}

// GetWaitTime returns recommended wait time before next action
func (rl *RateLimiter) GetWaitTime(actionType ActionType) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// If in cooldown, return remaining cooldown time
	if time.Now().Before(rl.cooldownUntil) {
		return time.Until(rl.cooldownUntil)
	}

	limit, exists := rl.limits[actionType]
	if !exists {
		return 0
	}

	history := rl.actionHistory[actionType]
	if len(history) == 0 {
		return 0
	}

	lastAction := history[len(history)-1]
	elapsed := time.Since(lastAction)

	if elapsed < limit.MinInterval {
		return limit.MinInterval - elapsed
	}

	return 0
}

// GetActionStats returns statistics for an action type
func (rl *RateLimiter) GetActionStats(actionType ActionType) map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit := rl.limits[actionType]
	history := rl.actionHistory[actionType]

	hourlyCount := rl.countActionsInWindow(history, time.Hour)
	dailyCount := rl.countActionsInWindow(history, 24*time.Hour)

	return map[string]interface{}{
		"action_type":      actionType,
		"hourly_count":     hourlyCount,
		"hourly_limit":     limit.HourlyMax,
		"hourly_remaining": limit.HourlyMax - hourlyCount,
		"daily_count":      dailyCount,
		"daily_limit":      limit.DailyMax,
		"daily_remaining":  limit.DailyMax - dailyCount,
		"in_cooldown":      time.Now().Before(rl.cooldownUntil),
		"cooldown_remaining": func() time.Duration {
			if time.Now().Before(rl.cooldownUntil) {
				return time.Until(rl.cooldownUntil)
			}
			return 0
		}(),
	}
}

// GetAllStats returns statistics for all action types
func (rl *RateLimiter) GetAllStats() map[ActionType]map[string]interface{} {
	stats := make(map[ActionType]map[string]interface{})

	for actionType := range rl.limits {
		stats[actionType] = rl.GetActionStats(actionType)
	}

	return stats
}

// ResetCooldown manually resets the cooldown period
func (rl *RateLimiter) ResetCooldown() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.cooldownUntil = time.Time{}
	rl.consecutiveActions = 0
}

// ResetDaily resets daily counters
func (rl *RateLimiter) ResetDaily() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Clear history older than 24 hours for all actions
	for actionType := range rl.actionHistory {
		rl.cleanHistory(actionType)
	}

	rl.dailyResetTime = time.Now().Add(24 * time.Hour)
}

// countActionsInWindow counts actions within a time window
func (rl *RateLimiter) countActionsInWindow(history []time.Time, window time.Duration) int {
	if len(history) == 0 {
		return 0
	}

	cutoff := time.Now().Add(-window)
	count := 0

	for i := len(history) - 1; i >= 0; i-- {
		if history[i].After(cutoff) {
			count++
		} else {
			break
		}
	}

	return count
}

// cleanHistory removes old entries from action history
func (rl *RateLimiter) cleanHistory(actionType ActionType) {
	history := rl.actionHistory[actionType]
	if len(history) == 0 {
		return
	}

	// Keep only last 24 hours of history
	cutoff := time.Now().Add(-24 * time.Hour)
	newHistory := []time.Time{}

	for _, t := range history {
		if t.After(cutoff) {
			newHistory = append(newHistory, t)
		}
	}

	rl.actionHistory[actionType] = newHistory
}

// resetTimers initializes reset timers
func (rl *RateLimiter) resetTimers() {
	now := time.Now()
	rl.hourlyResetTime = now.Add(time.Hour)
	rl.dailyResetTime = now.Add(24 * time.Hour)
}

// IsInCooldown checks if currently in cooldown
func (rl *RateLimiter) IsInCooldown() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return time.Now().Before(rl.cooldownUntil)
}

// GetCooldownRemaining returns remaining cooldown duration
func (rl *RateLimiter) GetCooldownRemaining() time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if time.Now().Before(rl.cooldownUntil) {
		return time.Until(rl.cooldownUntil)
	}
	return 0
}
