package security

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

// RateLimitManager provides comprehensive rate limiting and DDoS protection
type RateLimitManager struct {
	config           *RateLimitConfig
	globalLimiter    *rate.Limiter
	userLimiters     map[string]*UserRateLimiter
	ipLimiters       map[string]*IPRateLimiter
	bannedIPs        map[string]*BanInfo
	geoBlocker       *GeoBlocker
	statisticsCollector *RateLimitStatistics
	adaptiveEngine   *AdaptiveRateLimiting
	mu               sync.RWMutex
	cleanup          chan struct{}
}

// RateLimitConfig configures rate limiting
type RateLimitConfig struct {
	// Global limits
	GlobalEnabled      bool          `json:"global_enabled"`
	GlobalLimit        int           `json:"global_limit"`        // requests per second
	GlobalBurst        int           `json:"global_burst"`        // burst capacity
	
	// Per-user limits
	UserEnabled        bool          `json:"user_enabled"`
	UserLimit          int           `json:"user_limit"`          // requests per minute
	UserBurst          int           `json:"user_burst"`          // burst capacity
	
	// Per-IP limits
	IPEnabled          bool          `json:"ip_enabled"`
	IPLimit            int           `json:"ip_limit"`            // requests per minute
	IPBurst            int           `json:"ip_burst"`            // burst capacity
	
	// Ban configuration
	BanEnabled         bool          `json:"ban_enabled"`
	BanThreshold       int           `json:"ban_threshold"`       // violations before ban
	BanDuration        time.Duration `json:"ban_duration"`        // how long to ban
	BanCheckInterval   time.Duration `json:"ban_check_interval"`  // check interval
	
	// Whitelist/Blacklist
	WhitelistEnabled   bool          `json:"whitelist_enabled"`
	WhitelistedIPs     []string      `json:"whitelisted_ips"`
	WhitelistedCIDRs   []string      `json:"whitelisted_cidrs"`
	BlacklistedIPs     []string      `json:"blacklisted_ips"`
	BlacklistedCIDRs   []string      `json:"blacklisted_cidrs"`
	
	// Geo-blocking
	GeoBlockingEnabled bool          `json:"geo_blocking_enabled"`
	BlockedCountries   []string      `json:"blocked_countries"`
	AllowedCountries   []string      `json:"allowed_countries"`
	
	// Adaptive rate limiting
	AdaptiveEnabled    bool          `json:"adaptive_enabled"`
	AdaptiveThreshold  float64       `json:"adaptive_threshold"`  // system load threshold
	AdaptiveReduction  float64       `json:"adaptive_reduction"`  // reduction factor
	
	// DDoS protection
	DDoSEnabled        bool          `json:"ddos_enabled"`
	DDoSThreshold      int           `json:"ddos_threshold"`      // requests per second
	DDoSWindowSize     time.Duration `json:"ddos_window_size"`    // detection window
	DDoSAutoBlock      bool          `json:"ddos_auto_block"`     // auto-block on detection
	
	// Advanced features
	SlidingWindow      bool          `json:"sliding_window"`      // use sliding window algorithm
	TokenBucket        bool          `json:"token_bucket"`        // use token bucket algorithm
	DistributedMode    bool          `json:"distributed_mode"`    // distributed rate limiting
	RedisURL           string        `json:"redis_url"`           // for distributed mode
}

// UserRateLimiter tracks rate limiting for individual users
type UserRateLimiter struct {
	UserID         string         `json:"user_id"`
	Limiter        *rate.Limiter  `json:"-"`
	RequestCount   int64          `json:"request_count"`
	ViolationCount int            `json:"violation_count"`
	LastViolation  time.Time      `json:"last_violation"`
	CreatedAt      time.Time      `json:"created_at"`
	LastUsed       time.Time      `json:"last_used"`
}

// IPRateLimiter tracks rate limiting for IP addresses
type IPRateLimiter struct {
	IPAddress      string         `json:"ip_address"`
	Limiter        *rate.Limiter  `json:"-"`
	RequestCount   int64          `json:"request_count"`
	ViolationCount int            `json:"violation_count"`
	LastViolation  time.Time      `json:"last_violation"`
	Country        string         `json:"country,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	LastUsed       time.Time      `json:"last_used"`
}

// BanInfo represents information about a banned IP
type BanInfo struct {
	IPAddress    string        `json:"ip_address"`
	Reason       string        `json:"reason"`
	BannedAt     time.Time     `json:"banned_at"`
	ExpiresAt    time.Time     `json:"expires_at"`
	ViolationCount int         `json:"violation_count"`
	Country      string        `json:"country,omitempty"`
	UserAgent    string        `json:"user_agent,omitempty"`
	Permanent    bool          `json:"permanent"`
}

// GeoBlocker provides geographic IP blocking functionality
type GeoBlocker struct {
	enabled          bool
	blockedCountries map[string]bool
	allowedCountries map[string]bool
	defaultAction    string // allow, block
	geoDatabase      GeoDatabase
	mu               sync.RWMutex
}

// GeoDatabase interface for IP geolocation
type GeoDatabase interface {
	GetCountry(ip string) (string, error)
	GetLocation(ip string) (*GeoLocation, error)
}

// GeoLocation represents geographic location
type GeoLocation struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

// AdaptiveRateLimiting provides adaptive rate limiting based on system load
type AdaptiveRateLimiting struct {
	enabled          bool
	systemMonitor    *SystemMonitor
	thresholds       map[string]float64
	reductionFactors map[string]float64
	currentLimits    map[string]int
	mu               sync.RWMutex
}

// SystemMonitor monitors system performance metrics
type SystemMonitor struct {
	cpuUsage    float64
	memoryUsage float64
	diskIO      float64
	networkIO   float64
	lastUpdate  time.Time
}

// RateLimitStatistics tracks rate limiting statistics
type RateLimitStatistics struct {
	TotalRequests       int64                    `json:"total_requests"`
	AllowedRequests     int64                    `json:"allowed_requests"`
	BlockedRequests     int64                    `json:"blocked_requests"`
	BannedIPs           int64                    `json:"banned_ips"`
	ActiveUserLimiters  int64                    `json:"active_user_limiters"`
	ActiveIPLimiters    int64                    `json:"active_ip_limiters"`
	ViolationsByIP      map[string]int64         `json:"violations_by_ip"`
	ViolationsByUser    map[string]int64         `json:"violations_by_user"`
	RequestsByCountry   map[string]int64         `json:"requests_by_country"`
	DDoSAttemptsBlocked int64                    `json:"ddos_attempts_blocked"`
	LastUpdated         time.Time                `json:"last_updated"`
	mu                  sync.RWMutex
}

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed      bool          `json:"allowed"`
	Reason       string        `json:"reason"`
	RetryAfter   time.Duration `json:"retry_after"`
	Remaining    int           `json:"remaining"`
	ResetTime    time.Time     `json:"reset_time"`
	ViolationCount int         `json:"violation_count"`
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(config *RateLimitConfig) *RateLimitManager {
	rateLimitConfig := DefaultRateLimitConfig()
	
	// Override with provided config
	if config != nil {
		rateLimitConfig = config
	}

	rlm := &RateLimitManager{
		config:              rateLimitConfig,
		userLimiters:        make(map[string]*UserRateLimiter),
		ipLimiters:          make(map[string]*IPRateLimiter),
		bannedIPs:           make(map[string]*BanInfo),
		statisticsCollector: NewRateLimitStatistics(),
		cleanup:             make(chan struct{}),
	}

	// Initialize global rate limiter
	if rateLimitConfig.GlobalEnabled {
		rlm.globalLimiter = rate.NewLimiter(
			rate.Limit(rateLimitConfig.GlobalLimit),
			rateLimitConfig.GlobalBurst,
		)
	}

	// Initialize geo-blocker
	if rateLimitConfig.GeoBlockingEnabled {
		rlm.geoBlocker = NewGeoBlocker(rateLimitConfig)
	}

	// Initialize adaptive rate limiting
	if rateLimitConfig.AdaptiveEnabled {
		rlm.adaptiveEngine = NewAdaptiveRateLimiting(rateLimitConfig)
	}

	// Start background cleanup
	go rlm.startCleanup()

	// Start ban cleanup
	go rlm.startBanCleanup()

	log.Info().
		Bool("global_enabled", rateLimitConfig.GlobalEnabled).
		Bool("user_enabled", rateLimitConfig.UserEnabled).
		Bool("ip_enabled", rateLimitConfig.IPEnabled).
		Bool("geo_blocking", rateLimitConfig.GeoBlockingEnabled).
		Msg("Rate limit manager initialized")

	return rlm
}

// CheckRateLimit checks if a request should be rate limited
func (rlm *RateLimitManager) CheckRateLimit(c *gin.Context) bool {
	if !rlm.isAnyLimitEnabled() {
		return false
	}

	clientIP := c.ClientIP()
	userID := rlm.extractUserID(c)

	rlm.statisticsCollector.IncrementTotalRequests()

	// Check if IP is banned
	if rlm.config.BanEnabled {
		if banned, banInfo := rlm.isIPBanned(clientIP); banned {
			rlm.logRateLimitViolation(c, "ip_banned", banInfo)
			rlm.statisticsCollector.IncrementBlockedRequests()
			rlm.setRateLimitHeaders(c, &RateLimitResult{
				Allowed:    false,
				Reason:     "IP banned",
				RetryAfter: time.Until(banInfo.ExpiresAt),
			})
			return true
		}
	}

	// Check blacklist
	if rlm.isIPBlacklisted(clientIP) {
		rlm.logRateLimitViolation(c, "ip_blacklisted", nil)
		rlm.statisticsCollector.IncrementBlockedRequests()
		return true
	}

	// Skip rate limiting for whitelisted IPs
	if rlm.isIPWhitelisted(clientIP) {
		rlm.statisticsCollector.IncrementAllowedRequests()
		return false
	}

	// Check geo-blocking
	if rlm.config.GeoBlockingEnabled && rlm.geoBlocker != nil {
		if blocked, country := rlm.geoBlocker.IsBlocked(clientIP); blocked {
			rlm.logRateLimitViolation(c, "geo_blocked", map[string]interface{}{
				"country": country,
			})
			rlm.statisticsCollector.IncrementBlockedRequests()
			rlm.statisticsCollector.IncrementRequestsByCountry(country)
			return true
		}
	}

	// Check global rate limit
	if rlm.config.GlobalEnabled && rlm.globalLimiter != nil {
		if !rlm.globalLimiter.Allow() {
			rlm.logRateLimitViolation(c, "global_limit_exceeded", nil)
			rlm.statisticsCollector.IncrementBlockedRequests()
			rlm.setRateLimitHeaders(c, &RateLimitResult{
				Allowed:    false,
				Reason:     "Global rate limit exceeded",
				RetryAfter: time.Second,
			})
			return true
		}
	}

	// Check user rate limit
	if rlm.config.UserEnabled && userID != "" {
		if blocked, result := rlm.checkUserRateLimit(userID, c); blocked {
			rlm.logRateLimitViolation(c, "user_limit_exceeded", result)
			rlm.statisticsCollector.IncrementBlockedRequests()
			rlm.setRateLimitHeaders(c, result)
			return true
		}
	}

	// Check IP rate limit
	if rlm.config.IPEnabled {
		if blocked, result := rlm.checkIPRateLimit(clientIP, c); blocked {
			rlm.logRateLimitViolation(c, "ip_limit_exceeded", result)
			rlm.statisticsCollector.IncrementBlockedRequests()
			rlm.setRateLimitHeaders(c, result)
			
			// Check for ban threshold
			if rlm.config.BanEnabled && result.ViolationCount >= rlm.config.BanThreshold {
				rlm.banIP(clientIP, "rate_limit_violations", c.GetHeader("User-Agent"))
			}
			
			return true
		}
	}

	// Check for DDoS patterns
	if rlm.config.DDoSEnabled {
		if rlm.detectDDoS(clientIP, c) {
			rlm.logRateLimitViolation(c, "ddos_detected", nil)
			rlm.statisticsCollector.IncrementDDoSBlocked()
			
			if rlm.config.DDoSAutoBlock {
				rlm.banIP(clientIP, "ddos_attack", c.GetHeader("User-Agent"))
			}
			
			return true
		}
	}

	rlm.statisticsCollector.IncrementAllowedRequests()
	return false
}

// checkUserRateLimit checks user-specific rate limits
func (rlm *RateLimitManager) checkUserRateLimit(userID string, c *gin.Context) (bool, *RateLimitResult) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	userLimiter, exists := rlm.userLimiters[userID]
	if !exists {
		// Create new user limiter
		limiter := rate.NewLimiter(
			rate.Every(time.Minute/time.Duration(rlm.config.UserLimit)),
			rlm.config.UserBurst,
		)
		
		userLimiter = &UserRateLimiter{
			UserID:         userID,
			Limiter:        limiter,
			RequestCount:   0,
			ViolationCount: 0,
			CreatedAt:      time.Now(),
			LastUsed:       time.Now(),
		}
		
		rlm.userLimiters[userID] = userLimiter
	}

	userLimiter.RequestCount++
	userLimiter.LastUsed = time.Now()

	if !userLimiter.Limiter.Allow() {
		userLimiter.ViolationCount++
		userLimiter.LastViolation = time.Now()
		
		rlm.statisticsCollector.IncrementViolationsByUser(userID)
		
		return true, &RateLimitResult{
			Allowed:        false,
			Reason:         "User rate limit exceeded",
			RetryAfter:     time.Minute / time.Duration(rlm.config.UserLimit),
			ViolationCount: userLimiter.ViolationCount,
		}
	}

	return false, nil
}

// checkIPRateLimit checks IP-specific rate limits
func (rlm *RateLimitManager) checkIPRateLimit(clientIP string, c *gin.Context) (bool, *RateLimitResult) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	ipLimiter, exists := rlm.ipLimiters[clientIP]
	if !exists {
		// Create new IP limiter
		limiter := rate.NewLimiter(
			rate.Every(time.Minute/time.Duration(rlm.config.IPLimit)),
			rlm.config.IPBurst,
		)
		
		var country string
		if rlm.geoBlocker != nil {
			country, _ = rlm.geoBlocker.GetCountry(clientIP)
		}
		
		ipLimiter = &IPRateLimiter{
			IPAddress:      clientIP,
			Limiter:        limiter,
			RequestCount:   0,
			ViolationCount: 0,
			Country:        country,
			CreatedAt:      time.Now(),
			LastUsed:       time.Now(),
		}
		
		rlm.ipLimiters[clientIP] = ipLimiter
	}

	ipLimiter.RequestCount++
	ipLimiter.LastUsed = time.Now()

	if !ipLimiter.Limiter.Allow() {
		ipLimiter.ViolationCount++
		ipLimiter.LastViolation = time.Now()
		
		rlm.statisticsCollector.IncrementViolationsByIP(clientIP)
		
		return true, &RateLimitResult{
			Allowed:        false,
			Reason:         "IP rate limit exceeded",
			RetryAfter:     time.Minute / time.Duration(rlm.config.IPLimit),
			ViolationCount: ipLimiter.ViolationCount,
		}
	}

	return false, nil
}

// isIPBanned checks if an IP is currently banned
func (rlm *RateLimitManager) isIPBanned(ip string) (bool, *BanInfo) {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	banInfo, exists := rlm.bannedIPs[ip]
	if !exists {
		return false, nil
	}

	// Check if ban has expired (unless permanent)
	if !banInfo.Permanent && time.Now().After(banInfo.ExpiresAt) {
		delete(rlm.bannedIPs, ip)
		return false, nil
	}

	return true, banInfo
}

// banIP bans an IP address
func (rlm *RateLimitManager) banIP(ip, reason, userAgent string) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	var country string
	if rlm.geoBlocker != nil {
		country, _ = rlm.geoBlocker.GetCountry(ip)
	}

	banInfo := &BanInfo{
		IPAddress:      ip,
		Reason:         reason,
		BannedAt:       time.Now(),
		ExpiresAt:      time.Now().Add(rlm.config.BanDuration),
		ViolationCount: 1,
		Country:        country,
		UserAgent:      userAgent,
		Permanent:      false,
	}

	// Check if IP already banned and increment violation count
	if existing, exists := rlm.bannedIPs[ip]; exists {
		banInfo.ViolationCount = existing.ViolationCount + 1
		
		// Make ban permanent after repeated violations
		if banInfo.ViolationCount >= 5 {
			banInfo.Permanent = true
			banInfo.ExpiresAt = time.Time{}
		} else {
			// Exponential backoff for ban duration
			multiplier := time.Duration(banInfo.ViolationCount)
			banInfo.ExpiresAt = time.Now().Add(rlm.config.BanDuration * multiplier)
		}
	}

	rlm.bannedIPs[ip] = banInfo
	rlm.statisticsCollector.IncrementBannedIPs()

	log.Warn().
		Str("ip", ip).
		Str("reason", reason).
		Str("country", country).
		Time("expires_at", banInfo.ExpiresAt).
		Int("violation_count", banInfo.ViolationCount).
		Bool("permanent", banInfo.Permanent).
		Msg("IP address banned")
}

// UnbanIP removes an IP from the ban list
func (rlm *RateLimitManager) UnbanIP(ip string) error {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	if _, exists := rlm.bannedIPs[ip]; !exists {
		return fmt.Errorf("IP not found in ban list: %s", ip)
	}

	delete(rlm.bannedIPs, ip)

	log.Info().
		Str("ip", ip).
		Msg("IP address unbanned")

	return nil
}

// isIPWhitelisted checks if an IP is whitelisted
func (rlm *RateLimitManager) isIPWhitelisted(ip string) bool {
	if !rlm.config.WhitelistEnabled {
		return false
	}

	// Check exact IP matches
	for _, whiteIP := range rlm.config.WhitelistedIPs {
		if ip == whiteIP {
			return true
		}
	}

	// Check CIDR ranges
	for _, cidr := range rlm.config.WhitelistedCIDRs {
		if rlm.ipInCIDR(ip, cidr) {
			return true
		}
	}

	return false
}

// isIPBlacklisted checks if an IP is blacklisted
func (rlm *RateLimitManager) isIPBlacklisted(ip string) bool {
	// Check exact IP matches
	for _, blackIP := range rlm.config.BlacklistedIPs {
		if ip == blackIP {
			return true
		}
	}

	// Check CIDR ranges
	for _, cidr := range rlm.config.BlacklistedCIDRs {
		if rlm.ipInCIDR(ip, cidr) {
			return true
		}
	}

	return false
}

// ipInCIDR checks if an IP is in a CIDR range
func (rlm *RateLimitManager) ipInCIDR(ip, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return network.Contains(parsedIP)
}

// detectDDoS detects DDoS attack patterns
func (rlm *RateLimitManager) detectDDoS(ip string, c *gin.Context) bool {
	// Simplified DDoS detection - check request frequency
	if ipLimiter, exists := rlm.ipLimiters[ip]; exists {
		// Check if IP is making too many requests in a short time
		if ipLimiter.ViolationCount >= 3 {
			timeSinceLastViolation := time.Since(ipLimiter.LastViolation)
			if timeSinceLastViolation < rlm.config.DDoSWindowSize {
				return true
			}
		}
	}

	return false
}

// extractUserID extracts user ID from the request context
func (rlm *RateLimitManager) extractUserID(c *gin.Context) string {
	// Try to get user ID from different sources
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}

	// Try to get from JWT claims
	if claims, exists := c.Get("claims"); exists {
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if userID, exists := claimsMap["user_id"]; exists {
				if id, ok := userID.(string); ok {
					return id
				}
			}
		}
	}

	// Try to get from API key
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		// In real implementation, look up user ID from API key
		return fmt.Sprintf("api_key:%s", apiKey[:min(8, len(apiKey))])
	}

	return ""
}

// logRateLimitViolation logs rate limit violations
func (rlm *RateLimitManager) logRateLimitViolation(c *gin.Context, reason string, data interface{}) {
	log.Warn().
		Str("client_ip", c.ClientIP()).
		Str("method", c.Request.Method).
		Str("url", c.Request.URL.String()).
		Str("user_agent", c.GetHeader("User-Agent")).
		Str("reason", reason).
		Interface("data", data).
		Msg("Rate limit violation")
}

// setRateLimitHeaders sets HTTP headers with rate limit information
func (rlm *RateLimitManager) setRateLimitHeaders(c *gin.Context, result *RateLimitResult) {
	if result == nil {
		return
	}

	c.Header("X-RateLimit-Allowed", fmt.Sprintf("%t", result.Allowed))
	c.Header("X-RateLimit-Reason", result.Reason)
	
	if result.RetryAfter > 0 {
		c.Header("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
	}
	
	if result.Remaining > 0 {
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
	}
	
	if !result.ResetTime.IsZero() {
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))
	}
}

// isAnyLimitEnabled checks if any rate limiting is enabled
func (rlm *RateLimitManager) isAnyLimitEnabled() bool {
	return rlm.config.GlobalEnabled ||
		rlm.config.UserEnabled ||
		rlm.config.IPEnabled ||
		rlm.config.BanEnabled ||
		rlm.config.GeoBlockingEnabled
}

// startCleanup starts background cleanup of expired limiters
func (rlm *RateLimitManager) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rlm.cleanupExpiredLimiters()
		case <-rlm.cleanup:
			return
		}
	}
}

// cleanupExpiredLimiters removes expired user and IP limiters
func (rlm *RateLimitManager) cleanupExpiredLimiters() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	now := time.Now()
	expiry := 30 * time.Minute

	// Clean up user limiters
	for userID, limiter := range rlm.userLimiters {
		if now.Sub(limiter.LastUsed) > expiry {
			delete(rlm.userLimiters, userID)
		}
	}

	// Clean up IP limiters
	for ip, limiter := range rlm.ipLimiters {
		if now.Sub(limiter.LastUsed) > expiry {
			delete(rlm.ipLimiters, ip)
		}
	}

	log.Debug().
		Int("active_user_limiters", len(rlm.userLimiters)).
		Int("active_ip_limiters", len(rlm.ipLimiters)).
		Msg("Rate limiter cleanup completed")
}

// startBanCleanup starts background cleanup of expired bans
func (rlm *RateLimitManager) startBanCleanup() {
	ticker := time.NewTicker(rlm.config.BanCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rlm.cleanupExpiredBans()
		case <-rlm.cleanup:
			return
		}
	}
}

// cleanupExpiredBans removes expired IP bans
func (rlm *RateLimitManager) cleanupExpiredBans() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for ip, banInfo := range rlm.bannedIPs {
		if !banInfo.Permanent && now.After(banInfo.ExpiresAt) {
			delete(rlm.bannedIPs, ip)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		log.Info().
			Int("expired_bans", expiredCount).
			Int("active_bans", len(rlm.bannedIPs)).
			Msg("Expired IP bans cleaned up")
	}
}

// GetStatistics returns rate limiting statistics
func (rlm *RateLimitManager) GetStatistics() *RateLimitStatistics {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	rlm.statisticsCollector.mu.Lock()
	defer rlm.statisticsCollector.mu.Unlock()

	// Update current counts
	rlm.statisticsCollector.ActiveUserLimiters = int64(len(rlm.userLimiters))
	rlm.statisticsCollector.ActiveIPLimiters = int64(len(rlm.ipLimiters))
	rlm.statisticsCollector.BannedIPs = int64(len(rlm.bannedIPs))

	return rlm.statisticsCollector.GetSnapshot()
}

// GetBannedIPs returns a list of currently banned IPs
func (rlm *RateLimitManager) GetBannedIPs() map[string]*BanInfo {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	// Create a copy to avoid race conditions
	banned := make(map[string]*BanInfo)
	for ip, info := range rlm.bannedIPs {
		infoCopy := *info
		banned[ip] = &infoCopy
	}

	return banned
}

// Shutdown gracefully shuts down the rate limit manager
func (rlm *RateLimitManager) Shutdown(ctx context.Context) error {
	close(rlm.cleanup)
	return nil
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewGeoBlocker creates a new geo-blocker
func NewGeoBlocker(config *RateLimitConfig) *GeoBlocker {
	gb := &GeoBlocker{
		enabled:          config.GeoBlockingEnabled,
		blockedCountries: make(map[string]bool),
		allowedCountries: make(map[string]bool),
		defaultAction:    "allow",
	}

	// Populate blocked countries
	for _, country := range config.BlockedCountries {
		gb.blockedCountries[strings.ToUpper(country)] = true
	}

	// Populate allowed countries
	for _, country := range config.AllowedCountries {
		gb.allowedCountries[strings.ToUpper(country)] = true
	}

	return gb
}

// IsBlocked checks if an IP should be geo-blocked
func (gb *GeoBlocker) IsBlocked(ip string) (bool, string) {
	if !gb.enabled {
		return false, ""
	}

	country, err := gb.GetCountry(ip)
	if err != nil {
		// If we can't determine country, use default action
		return gb.defaultAction == "block", "unknown"
	}

	countryUpper := strings.ToUpper(country)

	// Check blocked countries first
	if _, blocked := gb.blockedCountries[countryUpper]; blocked {
		return true, country
	}

	// If allowed countries are specified, check if country is in the list
	if len(gb.allowedCountries) > 0 {
		if _, allowed := gb.allowedCountries[countryUpper]; !allowed {
			return true, country
		}
	}

	return false, country
}

// GetCountry gets country for an IP (simplified implementation)
func (gb *GeoBlocker) GetCountry(ip string) (string, error) {
	// Simplified implementation - in real implementation, use GeoIP database
	// This is just a placeholder that returns "US" for local IPs
	if ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "US", nil
	}
	
	// Return unknown for other IPs
	return "unknown", fmt.Errorf("country lookup not implemented")
}

// NewAdaptiveRateLimiting creates adaptive rate limiting engine
func NewAdaptiveRateLimiting(config *RateLimitConfig) *AdaptiveRateLimiting {
	return &AdaptiveRateLimiting{
		enabled:          config.AdaptiveEnabled,
		systemMonitor:    &SystemMonitor{},
		thresholds:       make(map[string]float64),
		reductionFactors: make(map[string]float64),
		currentLimits:    make(map[string]int),
	}
}

// NewRateLimitStatistics creates a new rate limit statistics collector
func NewRateLimitStatistics() *RateLimitStatistics {
	return &RateLimitStatistics{
		ViolationsByIP:      make(map[string]int64),
		ViolationsByUser:    make(map[string]int64),
		RequestsByCountry:   make(map[string]int64),
		LastUpdated:         time.Now(),
	}
}

// Statistical methods for RateLimitStatistics
func (rls *RateLimitStatistics) IncrementTotalRequests() {
	rls.mu.Lock()
	defer rls.mu.Unlock()
	rls.TotalRequests++
	rls.LastUpdated = time.Now()
}

func (rls *RateLimitStatistics) IncrementAllowedRequests() {
	rls.mu.Lock()
	defer rls.mu.Unlock()
	rls.AllowedRequests++
	rls.LastUpdated = time.Now()
}

func (rls *RateLimitStatistics) IncrementBlockedRequests() {
	rls.mu.Lock()
	defer rls.mu.Unlock()
	rls.BlockedRequests++
	rls.LastUpdated = time.Now()
}

func (rls *RateLimitStatistics) IncrementBannedIPs() {
	rls.mu.Lock()
	defer rls.mu.Unlock()
	rls.BannedIPs++
	rls.LastUpdated = time.Now()
}

func (rls *RateLimitStatistics) IncrementViolationsByIP(ip string) {
	rls.mu.Lock()
	defer rls.mu.Unlock()
	rls.ViolationsByIP[ip]++
	rls.LastUpdated = time.Now()
}

func (rls *RateLimitStatistics) IncrementViolationsByUser(userID string) {
	rls.mu.Lock()
	defer rls.mu.Unlock()
	rls.ViolationsByUser[userID]++
	rls.LastUpdated = time.Now()
}

func (rls *RateLimitStatistics) IncrementRequestsByCountry(country string) {
	rls.mu.Lock()
	defer rls.mu.Unlock()
	rls.RequestsByCountry[country]++
	rls.LastUpdated = time.Now()
}

func (rls *RateLimitStatistics) IncrementDDoSBlocked() {
	rls.mu.Lock()
	defer rls.mu.Unlock()
	rls.DDoSAttemptsBlocked++
	rls.LastUpdated = time.Now()
}

func (rls *RateLimitStatistics) GetSnapshot() *RateLimitStatistics {
	rls.mu.RLock()
	defer rls.mu.RUnlock()

	snapshot := &RateLimitStatistics{
		TotalRequests:       rls.TotalRequests,
		AllowedRequests:     rls.AllowedRequests,
		BlockedRequests:     rls.BlockedRequests,
		BannedIPs:           rls.BannedIPs,
		ActiveUserLimiters:  rls.ActiveUserLimiters,
		ActiveIPLimiters:    rls.ActiveIPLimiters,
		DDoSAttemptsBlocked: rls.DDoSAttemptsBlocked,
		LastUpdated:         rls.LastUpdated,
		ViolationsByIP:      make(map[string]int64),
		ViolationsByUser:    make(map[string]int64),
		RequestsByCountry:   make(map[string]int64),
	}

	// Copy maps
	for k, v := range rls.ViolationsByIP {
		snapshot.ViolationsByIP[k] = v
	}
	for k, v := range rls.ViolationsByUser {
		snapshot.ViolationsByUser[k] = v
	}
	for k, v := range rls.RequestsByCountry {
		snapshot.RequestsByCountry[k] = v
	}

	return snapshot
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		// Global limits
		GlobalEnabled: true,
		GlobalLimit:   10000, // 10k requests per second
		GlobalBurst:   100,
		
		// Per-user limits
		UserEnabled: true,
		UserLimit:   1000, // 1k requests per minute
		UserBurst:   50,
		
		// Per-IP limits
		IPEnabled: true,
		IPLimit:   100, // 100 requests per minute
		IPBurst:   20,
		
		// Ban configuration
		BanEnabled:       true,
		BanThreshold:     5,
		BanDuration:      15 * time.Minute,
		BanCheckInterval: time.Minute,
		
		// Whitelist/Blacklist
		WhitelistEnabled: false,
		WhitelistedIPs:   []string{},
		WhitelistedCIDRs: []string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		BlacklistedIPs:   []string{},
		BlacklistedCIDRs: []string{},
		
		// Geo-blocking
		GeoBlockingEnabled: false,
		BlockedCountries:   []string{},
		AllowedCountries:   []string{},
		
		// Adaptive rate limiting
		AdaptiveEnabled:   false,
		AdaptiveThreshold: 0.8,
		AdaptiveReduction: 0.5,
		
		// DDoS protection
		DDoSEnabled:    true,
		DDoSThreshold:  1000, // 1k requests per second
		DDoSWindowSize: 10 * time.Second,
		DDoSAutoBlock:  true,
		
		// Advanced features
		SlidingWindow:   false,
		TokenBucket:     true,
		DistributedMode: false,
	}
}