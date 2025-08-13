package auth

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// RateLimiter provides comprehensive rate limiting with multiple strategies
type RateLimiter struct {
	mu sync.RWMutex

	// Per-IP rate limiting
	ipLimits map[string]*IPLimit

	// Per-user rate limiting
	userLimits map[string]*UserLimit

	// Global rate limiting
	globalLimit *GlobalLimit

	// Configuration
	config *RateLimitConfig

	// Cleanup ticker
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Per-IP limits
	IPRequestsPerMinute int
	IPRequestsPerHour   int
	IPBurstSize         int

	// Per-user limits
	UserRequestsPerMinute int
	UserRequestsPerHour   int
	UserBurstSize         int

	// Global limits
	GlobalRequestsPerSecond int
	GlobalBurstSize         int

	// Elevated limits for authenticated users
	AuthenticatedMultiplier float64

	// Cleanup intervals
	CleanupInterval time.Duration

	// Blocking configuration
	BlockDuration time.Duration
	MaxViolations int

	// Whitelist/Blacklist
	WhitelistIPs []string
	BlacklistIPs []string
}

// IPLimit tracks rate limiting per IP address
type IPLimit struct {
	IP                 string
	RequestsThisMinute int
	RequestsThisHour   int
	MinuteWindow       int64
	HourWindow         int64
	LastRequest        time.Time
	Violations         int
	BlockedUntil       *time.Time
}

// UserLimit tracks rate limiting per user
type UserLimit struct {
	UserID             string
	RequestsThisMinute int
	RequestsThisHour   int
	MinuteWindow       int64
	HourWindow         int64
	LastRequest        time.Time
	IsAuthenticated    bool
	Violations         int
}

// GlobalLimit tracks global system rate limiting
type GlobalLimit struct {
	RequestsThisSecond int
	SecondWindow       int64
	LastRequest        time.Time
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed           bool
	Reason            string
	RetryAfter        time.Duration
	RequestsRemaining int
	ResetTime         time.Time
	WindowInfo        map[string]interface{}
}

// NewRateLimiter creates a new rate limiter with default configuration
func NewRateLimiter() *RateLimiter {
	config := &RateLimitConfig{
		IPRequestsPerMinute:     100,
		IPRequestsPerHour:       1000,
		IPBurstSize:             10,
		UserRequestsPerMinute:   200,
		UserRequestsPerHour:     2000,
		UserBurstSize:           20,
		GlobalRequestsPerSecond: 1000,
		GlobalBurstSize:         100,
		AuthenticatedMultiplier: 2.0,
		CleanupInterval:         5 * time.Minute,
		BlockDuration:           15 * time.Minute,
		MaxViolations:           5,
		WhitelistIPs:            []string{},
		BlacklistIPs:            []string{},
	}

	rl := &RateLimiter{
		ipLimits:    make(map[string]*IPLimit),
		userLimits:  make(map[string]*UserLimit),
		globalLimit: &GlobalLimit{},
		config:      config,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup routine
	rl.cleanupTicker = time.NewTicker(config.CleanupInterval)
	go rl.cleanup()

	return rl
}

// CheckRateLimit checks if a request should be allowed
func (rl *RateLimiter) CheckRateLimit(clientIP, userID string, isAuthenticated bool) *RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Check blacklist first
	if rl.isBlacklisted(clientIP) {
		return &RateLimitResult{
			Allowed:    false,
			Reason:     "IP address is blacklisted",
			RetryAfter: time.Hour, // Long retry for blacklisted IPs
		}
	}

	// Check whitelist (bypasses rate limiting)
	if rl.isWhitelisted(clientIP) {
		return &RateLimitResult{
			Allowed:           true,
			Reason:            "IP address is whitelisted",
			RequestsRemaining: 9999,
		}
	}

	// Check global rate limit first
	if !rl.checkGlobalLimit(now) {
		return &RateLimitResult{
			Allowed:    false,
			Reason:     "Global rate limit exceeded",
			RetryAfter: time.Second,
			ResetTime:  time.Unix((now.Unix()/60+1)*60, 0),
		}
	}

	// Check IP-based rate limiting
	ipResult := rl.checkIPLimit(clientIP, now, isAuthenticated)
	if !ipResult.Allowed {
		return ipResult
	}

	// Check user-based rate limiting if user is provided
	if userID != "" {
		userResult := rl.checkUserLimit(userID, now, isAuthenticated)
		if !userResult.Allowed {
			return userResult
		}

		// Return the more restrictive of the two limits
		if userResult.RequestsRemaining < ipResult.RequestsRemaining {
			return userResult
		}
	}

	return ipResult
}

// checkGlobalLimit checks the global system rate limit
func (rl *RateLimiter) checkGlobalLimit(now time.Time) bool {
	currentSecond := now.Unix()

	if rl.globalLimit.SecondWindow != currentSecond {
		rl.globalLimit.SecondWindow = currentSecond
		rl.globalLimit.RequestsThisSecond = 0
	}

	if rl.globalLimit.RequestsThisSecond >= rl.config.GlobalRequestsPerSecond {
		return false
	}

	rl.globalLimit.RequestsThisSecond++
	rl.globalLimit.LastRequest = now
	return true
}

// checkIPLimit checks per-IP rate limiting
func (rl *RateLimiter) checkIPLimit(clientIP string, now time.Time, isAuthenticated bool) *RateLimitResult {
	currentMinute := now.Unix() / 60
	currentHour := now.Unix() / 3600

	// Get or create IP limit entry
	ipLimit, exists := rl.ipLimits[clientIP]
	if !exists {
		ipLimit = &IPLimit{
			IP:           clientIP,
			MinuteWindow: currentMinute,
			HourWindow:   currentHour,
		}
		rl.ipLimits[clientIP] = ipLimit
	}

	// Check if IP is currently blocked
	if ipLimit.BlockedUntil != nil && now.Before(*ipLimit.BlockedUntil) {
		return &RateLimitResult{
			Allowed:    false,
			Reason:     "IP address is temporarily blocked due to rate limit violations",
			RetryAfter: ipLimit.BlockedUntil.Sub(now),
			ResetTime:  *ipLimit.BlockedUntil,
		}
	}

	// Reset counters if windows have changed
	if ipLimit.MinuteWindow != currentMinute {
		ipLimit.MinuteWindow = currentMinute
		ipLimit.RequestsThisMinute = 0
	}

	if ipLimit.HourWindow != currentHour {
		ipLimit.HourWindow = currentHour
		ipLimit.RequestsThisHour = 0
	}

	// Calculate effective limits (higher for authenticated users)
	minuteLimit := rl.config.IPRequestsPerMinute
	hourLimit := rl.config.IPRequestsPerHour

	if isAuthenticated {
		minuteLimit = int(float64(minuteLimit) * rl.config.AuthenticatedMultiplier)
		hourLimit = int(float64(hourLimit) * rl.config.AuthenticatedMultiplier)
	}

	// Check minute limit
	if ipLimit.RequestsThisMinute >= minuteLimit {
		rl.recordViolation(ipLimit, now)
		return &RateLimitResult{
			Allowed:           false,
			Reason:            "Per-IP minute rate limit exceeded",
			RetryAfter:        time.Duration((currentMinute+1)*60-now.Unix()) * time.Second,
			RequestsRemaining: 0,
			ResetTime:         time.Unix((currentMinute+1)*60, 0),
			WindowInfo: map[string]interface{}{
				"window": "minute",
				"limit":  minuteLimit,
				"used":   ipLimit.RequestsThisMinute,
			},
		}
	}

	// Check hour limit
	if ipLimit.RequestsThisHour >= hourLimit {
		rl.recordViolation(ipLimit, now)
		return &RateLimitResult{
			Allowed:           false,
			Reason:            "Per-IP hour rate limit exceeded",
			RetryAfter:        time.Duration((currentHour+1)*3600-now.Unix()) * time.Second,
			RequestsRemaining: 0,
			ResetTime:         time.Unix((currentHour+1)*3600, 0),
			WindowInfo: map[string]interface{}{
				"window": "hour",
				"limit":  hourLimit,
				"used":   ipLimit.RequestsThisHour,
			},
		}
	}

	// Increment counters
	ipLimit.RequestsThisMinute++
	ipLimit.RequestsThisHour++
	ipLimit.LastRequest = now

	// Clear blocked status if it has expired
	if ipLimit.BlockedUntil != nil && now.After(*ipLimit.BlockedUntil) {
		ipLimit.BlockedUntil = nil
		ipLimit.Violations = 0 // Reset violations after successful unblock
	}

	return &RateLimitResult{
		Allowed:           true,
		RequestsRemaining: minuteLimit - ipLimit.RequestsThisMinute,
		ResetTime:         time.Unix((currentMinute+1)*60, 0),
		WindowInfo: map[string]interface{}{
			"window": "minute",
			"limit":  minuteLimit,
			"used":   ipLimit.RequestsThisMinute,
		},
	}
}

// checkUserLimit checks per-user rate limiting
func (rl *RateLimiter) checkUserLimit(userID string, now time.Time, isAuthenticated bool) *RateLimitResult {
	currentMinute := now.Unix() / 60
	currentHour := now.Unix() / 3600

	// Get or create user limit entry
	userLimit, exists := rl.userLimits[userID]
	if !exists {
		userLimit = &UserLimit{
			UserID:          userID,
			MinuteWindow:    currentMinute,
			HourWindow:      currentHour,
			IsAuthenticated: isAuthenticated,
		}
		rl.userLimits[userID] = userLimit
	}

	// Reset counters if windows have changed
	if userLimit.MinuteWindow != currentMinute {
		userLimit.MinuteWindow = currentMinute
		userLimit.RequestsThisMinute = 0
	}

	if userLimit.HourWindow != currentHour {
		userLimit.HourWindow = currentHour
		userLimit.RequestsThisHour = 0
	}

	// Calculate effective limits
	minuteLimit := rl.config.UserRequestsPerMinute
	hourLimit := rl.config.UserRequestsPerHour

	if isAuthenticated {
		minuteLimit = int(float64(minuteLimit) * rl.config.AuthenticatedMultiplier)
		hourLimit = int(float64(hourLimit) * rl.config.AuthenticatedMultiplier)
	}

	// Check minute limit
	if userLimit.RequestsThisMinute >= minuteLimit {
		return &RateLimitResult{
			Allowed:           false,
			Reason:            "Per-user minute rate limit exceeded",
			RetryAfter:        time.Duration((currentMinute+1)*60-now.Unix()) * time.Second,
			RequestsRemaining: 0,
			ResetTime:         time.Unix((currentMinute+1)*60, 0),
			WindowInfo: map[string]interface{}{
				"window": "minute",
				"limit":  minuteLimit,
				"used":   userLimit.RequestsThisMinute,
			},
		}
	}

	// Check hour limit
	if userLimit.RequestsThisHour >= hourLimit {
		return &RateLimitResult{
			Allowed:           false,
			Reason:            "Per-user hour rate limit exceeded",
			RetryAfter:        time.Duration((currentHour+1)*3600-now.Unix()) * time.Second,
			RequestsRemaining: 0,
			ResetTime:         time.Unix((currentHour+1)*3600, 0),
			WindowInfo: map[string]interface{}{
				"window": "hour",
				"limit":  hourLimit,
				"used":   userLimit.RequestsThisHour,
			},
		}
	}

	// Increment counters
	userLimit.RequestsThisMinute++
	userLimit.RequestsThisHour++
	userLimit.LastRequest = now

	return &RateLimitResult{
		Allowed:           true,
		RequestsRemaining: minuteLimit - userLimit.RequestsThisMinute,
		ResetTime:         time.Unix((currentMinute+1)*60, 0),
		WindowInfo: map[string]interface{}{
			"window": "minute",
			"limit":  minuteLimit,
			"used":   userLimit.RequestsThisMinute,
		},
	}
}

// recordViolation records a rate limit violation and potentially blocks the IP
func (rl *RateLimiter) recordViolation(ipLimit *IPLimit, now time.Time) {
	ipLimit.Violations++

	if ipLimit.Violations >= rl.config.MaxViolations {
		blockUntil := now.Add(rl.config.BlockDuration)
		ipLimit.BlockedUntil = &blockUntil

		// Log the blocking event (in production, send to monitoring system)
		fmt.Printf("IP %s blocked until %v due to %d violations\\n",
			ipLimit.IP, blockUntil, ipLimit.Violations)
	}
}

// isWhitelisted checks if an IP is in the whitelist
func (rl *RateLimiter) isWhitelisted(clientIP string) bool {
	for _, whiteIP := range rl.config.WhitelistIPs {
		if rl.matchIP(clientIP, whiteIP) {
			return true
		}
	}
	return false
}

// isBlacklisted checks if an IP is in the blacklist
func (rl *RateLimiter) isBlacklisted(clientIP string) bool {
	for _, blackIP := range rl.config.BlacklistIPs {
		if rl.matchIP(clientIP, blackIP) {
			return true
		}
	}
	return false
}

// matchIP checks if an IP matches a pattern (supports CIDR notation)
func (rl *RateLimiter) matchIP(clientIP, pattern string) bool {
	// Try exact match first
	if clientIP == pattern {
		return true
	}

	// Try CIDR match
	_, network, err := net.ParseCIDR(pattern)
	if err != nil {
		return false
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	return network.Contains(ip)
}

// cleanup removes old entries from the rate limiter
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.performCleanup()
		case <-rl.stopCleanup:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// performCleanup removes stale entries
func (rl *RateLimiter) performCleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Hour) // Remove entries older than 1 hour

	// Clean up IP limits
	for ip, limit := range rl.ipLimits {
		if limit.LastRequest.Before(cutoff) &&
			(limit.BlockedUntil == nil || now.After(*limit.BlockedUntil)) {
			delete(rl.ipLimits, ip)
		}
	}

	// Clean up user limits
	for userID, limit := range rl.userLimits {
		if limit.LastRequest.Before(cutoff) {
			delete(rl.userLimits, userID)
		}
	}
}

// GetStats returns current rate limiting statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	blockedIPs := 0
	for _, limit := range rl.ipLimits {
		if limit.BlockedUntil != nil && time.Now().Before(*limit.BlockedUntil) {
			blockedIPs++
		}
	}

	return map[string]interface{}{
		"tracked_ips":         len(rl.ipLimits),
		"tracked_users":       len(rl.userLimits),
		"blocked_ips":         blockedIPs,
		"global_requests_sec": rl.globalLimit.RequestsThisSecond,
		"config": map[string]interface{}{
			"ip_requests_per_minute":   rl.config.IPRequestsPerMinute,
			"user_requests_per_minute": rl.config.UserRequestsPerMinute,
			"global_requests_per_sec":  rl.config.GlobalRequestsPerSecond,
		},
	}
}

// Close gracefully shuts down the rate limiter
func (rl *RateLimiter) Close() {
	close(rl.stopCleanup)
}
