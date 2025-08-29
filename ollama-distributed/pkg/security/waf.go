package security

import (
	_ "bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// WAFManager provides Web Application Firewall functionality
type WAFManager struct {
	config          *WAFConfig
	rules           map[string]*WAFRule
	ruleEngine      *RuleEngine
	enabled         bool
	logBlocked      bool
	statisticsCollector *WAFStatistics
	customRules     []*CustomWAFRule
	mu              sync.RWMutex
}

// WAFConfig configures the WAF
type WAFConfig struct {
	Enabled            bool              `json:"enabled"`
	Mode               string            `json:"mode"` // detection, prevention
	LogBlocked         bool              `json:"log_blocked"`
	BlockByDefault     bool              `json:"block_by_default"`
	EnableOWASPCRS     bool              `json:"enable_owasp_crs"`
	CustomRulesPath    string            `json:"custom_rules_path"`
	MaxRequestSize     int64             `json:"max_request_size"`
	MaxHeaderSize      int               `json:"max_header_size"`
	MaxQueryParams     int               `json:"max_query_params"`
	MaxPostParams      int               `json:"max_post_params"`
	BlockedResponse    string            `json:"blocked_response"`
	AllowedMethods     []string          `json:"allowed_methods"`
	AllowedExtensions  []string          `json:"allowed_extensions"`
	BlockedExtensions  []string          `json:"blocked_extensions"`
	GeoBlocking        *GeoBlockingConfig `json:"geo_blocking"`
	RateLimiting       *WAFRateLimitConfig   `json:"rate_limiting"`
}

// GeoBlockingConfig configures geographic blocking
type GeoBlockingConfig struct {
	Enabled         bool     `json:"enabled"`
	BlockedCountries []string `json:"blocked_countries"`
	AllowedCountries []string `json:"allowed_countries"`
	DefaultAction   string   `json:"default_action"` // allow, block
}

// WAFRateLimitConfig configures rate limiting within WAF
type WAFRateLimitConfig struct {
	Enabled        bool          `json:"enabled"`
	RequestsPerMin int           `json:"requests_per_minute"`
	BurstSize      int           `json:"burst_size"`
	BanDuration    time.Duration `json:"ban_duration"`
}

// WAFRule represents a WAF security rule
type WAFRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        WAFRuleType       `json:"type"`
	Pattern     string            `json:"pattern"`
	Regex       *regexp.Regexp    `json:"-"`
	Action      WAFAction         `json:"action"`
	Severity    SeverityLevel     `json:"severity"`
	Categories  []string          `json:"categories"`
	Targets     []WAFTarget       `json:"targets"`
	Enabled     bool              `json:"enabled"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CustomWAFRule represents a user-defined WAF rule
type CustomWAFRule struct {
	*WAFRule
	Priority int    `json:"priority"`
	Owner    string `json:"owner"`
}

// WAFRuleType defines the type of WAF rule
type WAFRuleType string

const (
	WAFRuleTypeSignature  WAFRuleType = "signature"
	WAFRuleTypeRegex      WAFRuleType = "regex"
	WAFRuleTypeBehavioral WAFRuleType = "behavioral"
	WAFRuleTypeAnomaly    WAFRuleType = "anomaly"
	WAFRuleTypeReputation WAFRuleType = "reputation"
)

// WAFAction defines the action to take when a rule matches
type WAFAction string

const (
	WAFActionAllow     WAFAction = "allow"
	WAFActionLog       WAFAction = "log"
	WAFActionBlock     WAFAction = "block"
	WAFActionChallenge WAFAction = "challenge"
	WAFActionRedirect  WAFAction = "redirect"
	WAFActionDrop      WAFAction = "drop"
)

// WAFTarget defines what part of the request to inspect
type WAFTarget string

const (
	WAFTargetURL         WAFTarget = "url"
	WAFTargetQueryString WAFTarget = "query_string"
	WAFTargetHeaders     WAFTarget = "headers"
	WAFTargetBody        WAFTarget = "body"
	WAFTargetCookies     WAFTarget = "cookies"
	WAFTargetUserAgent   WAFTarget = "user_agent"
	WAFTargetReferer     WAFTarget = "referer"
	WAFTargetMethod      WAFTarget = "method"
	WAFTargetIPAddress   WAFTarget = "ip_address"
)

// RuleEngine processes WAF rules
type RuleEngine struct {
	compiledRules map[string]*CompiledRule
	ruleChains    []*RuleChain
	preprocessors []Preprocessor
	mu            sync.RWMutex
}

// CompiledRule represents a compiled WAF rule for efficient matching
type CompiledRule struct {
	Rule     *WAFRule
	Matchers []Matcher
}

// RuleChain represents a chain of related rules
type RuleChain struct {
	ID    string
	Rules []*WAFRule
	Logic string // and, or
}

// Matcher interface for different matching strategies
type Matcher interface {
	Match(request *RequestContext) (bool, *MatchResult)
	GetType() string
}

// MatchResult contains the result of a rule match
type MatchResult struct {
	Matched     bool              `json:"matched"`
	Rule        *WAFRule          `json:"rule"`
	Target      WAFTarget         `json:"target"`
	Value       string            `json:"value"`
	Evidence    string            `json:"evidence"`
	Confidence  float64           `json:"confidence"`
	Metadata    map[string]string `json:"metadata"`
}

// RequestContext contains request information for WAF processing
type RequestContext struct {
	Request      *http.Request
	URL          *url.URL
	Method       string
	Headers      map[string][]string
	QueryParams  map[string][]string
	PostParams   map[string][]string
	Body         []byte
	ClientIP     string
	UserAgent    string
	Referer      string
	Cookies      []*http.Cookie
	ContentType  string
	ContentLength int64
}

// WAFStatistics tracks WAF performance and blocking statistics
type WAFStatistics struct {
	RequestsProcessed   int64                    `json:"requests_processed"`
	RequestsBlocked     int64                    `json:"requests_blocked"`
	RequestsAllowed     int64                    `json:"requests_allowed"`
	RuleMatches         map[string]int64         `json:"rule_matches"`
	CategoryMatches     map[string]int64         `json:"category_matches"`
	BlockedIPs          map[string]int64         `json:"blocked_ips"`
	TopAttackTypes      map[string]int64         `json:"top_attack_types"`
	ProcessingTime      []time.Duration          `json:"-"`
	AverageProcessingTime time.Duration          `json:"average_processing_time"`
	LastUpdated         time.Time                `json:"last_updated"`
	mu                  sync.RWMutex
}

// Preprocessor interface for request preprocessing
type Preprocessor interface {
	Process(ctx *RequestContext) error
	GetName() string
}

// NewWAFManager creates a new WAF manager
func NewWAFManager(rules []WAFRule) *WAFManager {
	config := DefaultWAFConfig()
	
	waf := &WAFManager{
		config:              config,
		rules:               make(map[string]*WAFRule),
		enabled:             config.Enabled,
		logBlocked:          config.LogBlocked,
		statisticsCollector: NewWAFStatistics(),
		customRules:         make([]*CustomWAFRule, 0),
	}

	// Initialize rule engine
	waf.ruleEngine = NewRuleEngine()

	// Load default OWASP CRS rules if enabled
	if config.EnableOWASPCRS {
		waf.loadOWASPCRS()
	}

	// Load provided rules
	for _, rule := range rules {
		waf.AddRule(&rule)
	}

	// Load custom rules from file if specified
	if config.CustomRulesPath != "" {
		if err := waf.loadCustomRules(config.CustomRulesPath); err != nil {
			log.Error().Err(err).Msg("Failed to load custom WAF rules")
		}
	}

	log.Info().
		Int("rules_loaded", len(waf.rules)).
		Bool("owasp_crs", config.EnableOWASPCRS).
		Msg("WAF manager initialized")

	return waf
}

// AddRule adds a new WAF rule
func (waf *WAFManager) AddRule(rule *WAFRule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	waf.mu.Lock()
	defer waf.mu.Unlock()

	// Compile regex if pattern is provided
	if rule.Pattern != "" {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
		rule.Regex = regex
	}

	rule.UpdatedAt = time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}

	waf.rules[rule.ID] = rule

	// Add to rule engine
	return waf.ruleEngine.AddRule(rule)
}

// RemoveRule removes a WAF rule
func (waf *WAFManager) RemoveRule(ruleID string) error {
	waf.mu.Lock()
	defer waf.mu.Unlock()

	if _, exists := waf.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	delete(waf.rules, ruleID)
	return waf.ruleEngine.RemoveRule(ruleID)
}

// CheckRequest processes a request through the WAF
func (waf *WAFManager) CheckRequest(r *http.Request) (bool, *WAFRule) {
	if !waf.enabled {
		return false, nil
	}

	start := time.Now()
	defer func() {
		waf.statisticsCollector.AddProcessingTime(time.Since(start))
	}()

	// Create request context
	ctx, err := waf.createRequestContext(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create request context")
		return false, nil
	}

	waf.statisticsCollector.IncrementProcessed()

	// Run preprocessors
	for _, preprocessor := range waf.ruleEngine.preprocessors {
		if err := preprocessor.Process(ctx); err != nil {
			log.Error().
				Err(err).
				Str("preprocessor", preprocessor.GetName()).
				Msg("Preprocessor failed")
		}
	}

	// Check basic request limits
	if blocked, rule := waf.checkBasicLimits(ctx); blocked {
		waf.handleBlocked(ctx, rule, "basic_limits")
		return true, rule
	}

	// Check allowed/blocked methods
	if blocked, rule := waf.checkMethods(ctx); blocked {
		waf.handleBlocked(ctx, rule, "method_check")
		return true, rule
	}

	// Check file extensions
	if blocked, rule := waf.checkExtensions(ctx); blocked {
		waf.handleBlocked(ctx, rule, "extension_check")
		return true, rule
	}

	// Process rules through the rule engine
	matches, err := waf.ruleEngine.ProcessRequest(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Rule engine processing failed")
		return false, nil
	}

	// Evaluate matches
	for _, match := range matches {
		if match.Matched {
			action := waf.evaluateAction(match)
			
			switch action {
			case WAFActionBlock, WAFActionDrop:
				waf.handleBlocked(ctx, match.Rule, "rule_match")
				return true, match.Rule
			case WAFActionChallenge:
				// Implement challenge logic (CAPTCHA, etc.)
				waf.handleChallenge(ctx, match.Rule)
				return true, match.Rule
			case WAFActionLog:
				waf.logMatch(ctx, match)
			}
		}
	}

	waf.statisticsCollector.IncrementAllowed()
	return false, nil
}

// createRequestContext creates a request context from HTTP request
func (waf *WAFManager) createRequestContext(r *http.Request) (*RequestContext, error) {
	// Parse URL
	parsedURL, err := url.Parse(r.URL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Parse query parameters
	queryParams := make(map[string][]string)
	for k, v := range parsedURL.Query() {
		queryParams[k] = v
	}

	// Parse POST parameters
	postParams := make(map[string][]string)
	if r.Method == "POST" && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err == nil {
			for k, v := range r.PostForm {
				postParams[k] = v
			}
		}
	}

	// Read body (with size limit)
	var body []byte
	if r.Body != nil && r.ContentLength > 0 && r.ContentLength <= waf.config.MaxRequestSize {
		body, err = io.ReadAll(io.LimitReader(r.Body, waf.config.MaxRequestSize))
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	ctx := &RequestContext{
		Request:       r,
		URL:           parsedURL,
		Method:        r.Method,
		Headers:       r.Header,
		QueryParams:   queryParams,
		PostParams:    postParams,
		Body:          body,
		ClientIP:      r.RemoteAddr,
		UserAgent:     r.UserAgent(),
		Referer:       r.Referer(),
		Cookies:       r.Cookies(),
		ContentType:   r.Header.Get("Content-Type"),
		ContentLength: r.ContentLength,
	}

	return ctx, nil
}

// checkBasicLimits checks basic request limits
func (waf *WAFManager) checkBasicLimits(ctx *RequestContext) (bool, *WAFRule) {
	// Check request size
	if ctx.ContentLength > waf.config.MaxRequestSize {
		return true, &WAFRule{
			ID:       "basic_limit_request_size",
			Name:     "Request Size Limit",
			Action:   WAFActionBlock,
			Severity: SeverityMedium,
		}
	}

	// Check header count and size
	headerCount := 0
	headerSize := 0
	for name, values := range ctx.Headers {
		headerCount++
		headerSize += len(name)
		for _, value := range values {
			headerSize += len(value)
		}
	}

	if headerSize > waf.config.MaxHeaderSize {
		return true, &WAFRule{
			ID:       "basic_limit_header_size",
			Name:     "Header Size Limit",
			Action:   WAFActionBlock,
			Severity: SeverityMedium,
		}
	}

	// Check query parameter count
	if len(ctx.QueryParams) > waf.config.MaxQueryParams {
		return true, &WAFRule{
			ID:       "basic_limit_query_params",
			Name:     "Query Parameter Count Limit",
			Action:   WAFActionBlock,
			Severity: SeverityLow,
		}
	}

	// Check POST parameter count
	if len(ctx.PostParams) > waf.config.MaxPostParams {
		return true, &WAFRule{
			ID:       "basic_limit_post_params",
			Name:     "POST Parameter Count Limit",
			Action:   WAFActionBlock,
			Severity: SeverityLow,
		}
	}

	return false, nil
}

// checkMethods checks allowed/blocked HTTP methods
func (waf *WAFManager) checkMethods(ctx *RequestContext) (bool, *WAFRule) {
	if len(waf.config.AllowedMethods) > 0 {
		allowed := false
		for _, method := range waf.config.AllowedMethods {
			if strings.EqualFold(ctx.Method, method) {
				allowed = true
				break
			}
		}
		
		if !allowed {
			return true, &WAFRule{
				ID:       "method_not_allowed",
				Name:     "HTTP Method Not Allowed",
				Action:   WAFActionBlock,
				Severity: SeverityMedium,
			}
		}
	}

	return false, nil
}

// checkExtensions checks file extension restrictions
func (waf *WAFManager) checkExtensions(ctx *RequestContext) (bool, *WAFRule) {
	path := ctx.URL.Path
	if path == "" || path == "/" {
		return false, nil
	}

	// Extract file extension
	lastDot := strings.LastIndex(path, ".")
	if lastDot == -1 {
		return false, nil
	}
	
	ext := strings.ToLower(path[lastDot+1:])

	// Check blocked extensions first
	for _, blockedExt := range waf.config.BlockedExtensions {
		if ext == strings.ToLower(blockedExt) {
			return true, &WAFRule{
				ID:       "blocked_extension",
				Name:     "Blocked File Extension",
				Action:   WAFActionBlock,
				Severity: SeverityHigh,
			}
		}
	}

	// Check allowed extensions if specified
	if len(waf.config.AllowedExtensions) > 0 {
		allowed := false
		for _, allowedExt := range waf.config.AllowedExtensions {
			if ext == strings.ToLower(allowedExt) {
				allowed = true
				break
			}
		}
		
		if !allowed {
			return true, &WAFRule{
				ID:       "extension_not_allowed",
				Name:     "File Extension Not Allowed",
				Action:   WAFActionBlock,
				Severity: SeverityMedium,
			}
		}
	}

	return false, nil
}

// evaluateAction determines the final action based on rule and configuration
func (waf *WAFManager) evaluateAction(match *MatchResult) WAFAction {
	// In detection mode, convert blocking actions to logging
	if waf.config.Mode == "detection" && match.Rule.Action == WAFActionBlock {
		return WAFActionLog
	}

	return match.Rule.Action
}

// handleBlocked handles blocked requests
func (waf *WAFManager) handleBlocked(ctx *RequestContext, rule *WAFRule, reason string) {
	waf.statisticsCollector.IncrementBlocked()
	waf.statisticsCollector.AddRuleMatch(rule.ID)
	waf.statisticsCollector.AddBlockedIP(ctx.ClientIP)

	if waf.logBlocked {
		log.Warn().
			Str("client_ip", ctx.ClientIP).
			Str("method", ctx.Method).
			Str("url", ctx.URL.String()).
			Str("user_agent", ctx.UserAgent).
			Str("rule_id", rule.ID).
			Str("rule_name", rule.Name).
			Str("reason", reason).
			Msg("WAF blocked request")
	}
}

// handleChallenge handles challenge requests
func (waf *WAFManager) handleChallenge(ctx *RequestContext, rule *WAFRule) {
	log.Info().
		Str("client_ip", ctx.ClientIP).
		Str("rule_id", rule.ID).
		Msg("WAF challenge issued")
}

// logMatch logs rule matches
func (waf *WAFManager) logMatch(ctx *RequestContext, match *MatchResult) {
	waf.statisticsCollector.AddRuleMatch(match.Rule.ID)

	log.Info().
		Str("client_ip", ctx.ClientIP).
		Str("method", ctx.Method).
		Str("url", ctx.URL.String()).
		Str("rule_id", match.Rule.ID).
		Str("rule_name", match.Rule.Name).
		Str("target", string(match.Target)).
		Float64("confidence", match.Confidence).
		Msg("WAF rule match (logged)")
}

// loadOWASPCRS loads OWASP Core Rule Set
func (waf *WAFManager) loadOWASPCRS() {
	owaspRules := []*WAFRule{
		{
			ID:          "owasp_crs_941100",
			Name:        "XSS Attack Detected",
			Description: "Cross-site Scripting (XSS) Attack",
			Type:        WAFRuleTypeSignature,
			Pattern:     `(?i)(<script[^>]*>.*?</script>|javascript:|vbscript:|onload=|onerror=)`,
			Action:      WAFActionBlock,
			Severity:    SeverityHigh,
			Categories:  []string{"xss", "injection"},
			Targets:     []WAFTarget{WAFTargetQueryString, WAFTargetBody, WAFTargetHeaders},
			Enabled:     true,
		},
		{
			ID:          "owasp_crs_942100",
			Name:        "SQL Injection Attack",
			Description: "SQL Injection Attack Detected",
			Type:        WAFRuleTypeSignature,
			Pattern:     `(?i)(union.*select|insert.*into|delete.*from|drop.*table|create.*table|alter.*table)`,
			Action:      WAFActionBlock,
			Severity:    SeverityCritical,
			Categories:  []string{"sqli", "injection"},
			Targets:     []WAFTarget{WAFTargetQueryString, WAFTargetBody, WAFTargetHeaders},
			Enabled:     true,
		},
		{
			ID:          "owasp_crs_913100",
			Name:        "Path Traversal Attack",
			Description: "Path Traversal/Directory Traversal Attack",
			Type:        WAFRuleTypeSignature,
			Pattern:     `(?i)(\.\.\/|\.\.\\|\/etc\/|\/proc\/|\/dev\/|c:\\)`,
			Action:      WAFActionBlock,
			Severity:    SeverityHigh,
			Categories:  []string{"path_traversal", "lfi"},
			Targets:     []WAFTarget{WAFTargetURL, WAFTargetQueryString, WAFTargetBody},
			Enabled:     true,
		},
		{
			ID:          "owasp_crs_920100",
			Name:        "HTTP Protocol Violation",
			Description: "Invalid HTTP Request",
			Type:        WAFRuleTypeBehavioral,
			Pattern:     ``,
			Action:      WAFActionBlock,
			Severity:    SeverityMedium,
			Categories:  []string{"protocol", "violation"},
			Targets:     []WAFTarget{WAFTargetHeaders},
			Enabled:     true,
		},
	}

	for _, rule := range owaspRules {
		if err := waf.AddRule(rule); err != nil {
			log.Error().
				Err(err).
				Str("rule_id", rule.ID).
				Msg("Failed to add OWASP CRS rule")
		}
	}

	log.Info().
		Int("rules_loaded", len(owaspRules)).
		Msg("OWASP CRS rules loaded")
}

// loadCustomRules loads custom rules from file
func (waf *WAFManager) loadCustomRules(filePath string) error {
	// Implementation would load rules from YAML/JSON file
	log.Info().
		Str("file_path", filePath).
		Msg("Loading custom WAF rules")
	
	return nil
}

// GetStatistics returns WAF statistics
func (waf *WAFManager) GetStatistics() *WAFStatistics {
	return waf.statisticsCollector.GetSnapshot()
}

// Middleware returns a Gin middleware for WAF protection
func (waf *WAFManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !waf.enabled {
			c.Next()
			return
		}

		blocked, rule := waf.CheckRequest(c.Request)
		if blocked {
			response := waf.config.BlockedResponse
			if response == "" {
				response = "Request blocked by Web Application Firewall"
			}

			// Log blocked request details
			log.Warn().
				Str("client_ip", c.ClientIP()).
				Str("method", c.Request.Method).
				Str("url", c.Request.URL.String()).
				Str("rule", rule.ID).
				Msg("WAF blocked request in middleware")

			c.JSON(http.StatusForbidden, gin.H{
				"error": response,
				"rule":  rule.ID,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		compiledRules: make(map[string]*CompiledRule),
		ruleChains:    make([]*RuleChain, 0),
		preprocessors: make([]Preprocessor, 0),
	}
}

// AddRule adds a rule to the engine
func (re *RuleEngine) AddRule(rule *WAFRule) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	compiledRule := &CompiledRule{
		Rule:     rule,
		Matchers: make([]Matcher, 0),
	}

	// Create matchers based on rule type and targets
	for _, target := range rule.Targets {
		var matcher Matcher
		switch rule.Type {
		case WAFRuleTypeSignature:
			matcher = NewSignatureMatcher(rule, target)
		case WAFRuleTypeRegex:
			matcher = NewRegexMatcher(rule, target)
		case WAFRuleTypeBehavioral:
			matcher = NewBehavioralMatcher(rule, target)
		default:
			matcher = NewRegexMatcher(rule, target)
		}
		
		if matcher != nil {
			compiledRule.Matchers = append(compiledRule.Matchers, matcher)
		}
	}

	re.compiledRules[rule.ID] = compiledRule
	return nil
}

// RemoveRule removes a rule from the engine
func (re *RuleEngine) RemoveRule(ruleID string) error {
	re.mu.Lock()
	defer re.mu.Unlock()

	delete(re.compiledRules, ruleID)
	return nil
}

// ProcessRequest processes a request through the rule engine
func (re *RuleEngine) ProcessRequest(ctx *RequestContext) ([]*MatchResult, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	results := make([]*MatchResult, 0)

	for _, compiledRule := range re.compiledRules {
		if !compiledRule.Rule.Enabled {
			continue
		}

		for _, matcher := range compiledRule.Matchers {
			matched, result := matcher.Match(ctx)
			if matched && result != nil {
				results = append(results, result)
				
				// Early exit for blocking rules in prevention mode
				if compiledRule.Rule.Action == WAFActionBlock {
					break
				}
			}
		}
	}

	return results, nil
}

// NewWAFStatistics creates a new WAF statistics collector
func NewWAFStatistics() *WAFStatistics {
	return &WAFStatistics{
		RuleMatches:      make(map[string]int64),
		CategoryMatches:  make(map[string]int64),
		BlockedIPs:       make(map[string]int64),
		TopAttackTypes:   make(map[string]int64),
		ProcessingTime:   make([]time.Duration, 0, 1000),
		LastUpdated:      time.Now(),
	}
}

// Statistical methods for WAFStatistics
func (ws *WAFStatistics) IncrementProcessed() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.RequestsProcessed++
	ws.LastUpdated = time.Now()
}

func (ws *WAFStatistics) IncrementBlocked() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.RequestsBlocked++
	ws.LastUpdated = time.Now()
}

func (ws *WAFStatistics) IncrementAllowed() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.RequestsAllowed++
	ws.LastUpdated = time.Now()
}

func (ws *WAFStatistics) AddRuleMatch(ruleID string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.RuleMatches[ruleID]++
	ws.LastUpdated = time.Now()
}

func (ws *WAFStatistics) AddBlockedIP(ip string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.BlockedIPs[ip]++
	ws.LastUpdated = time.Now()
}

func (ws *WAFStatistics) AddProcessingTime(duration time.Duration) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	
	ws.ProcessingTime = append(ws.ProcessingTime, duration)
	
	// Keep only last 1000 measurements
	if len(ws.ProcessingTime) > 1000 {
		ws.ProcessingTime = ws.ProcessingTime[1:]
	}
	
	// Recalculate average
	var total time.Duration
	for _, d := range ws.ProcessingTime {
		total += d
	}
	ws.AverageProcessingTime = total / time.Duration(len(ws.ProcessingTime))
	ws.LastUpdated = time.Now()
}

func (ws *WAFStatistics) GetSnapshot() *WAFStatistics {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	snapshot := &WAFStatistics{
		RequestsProcessed:     ws.RequestsProcessed,
		RequestsBlocked:       ws.RequestsBlocked,
		RequestsAllowed:       ws.RequestsAllowed,
		AverageProcessingTime: ws.AverageProcessingTime,
		LastUpdated:           ws.LastUpdated,
		RuleMatches:           make(map[string]int64),
		CategoryMatches:       make(map[string]int64),
		BlockedIPs:            make(map[string]int64),
		TopAttackTypes:        make(map[string]int64),
	}

	// Copy maps
	for k, v := range ws.RuleMatches {
		snapshot.RuleMatches[k] = v
	}
	for k, v := range ws.CategoryMatches {
		snapshot.CategoryMatches[k] = v
	}
	for k, v := range ws.BlockedIPs {
		snapshot.BlockedIPs[k] = v
	}
	for k, v := range ws.TopAttackTypes {
		snapshot.TopAttackTypes[k] = v
	}

	return snapshot
}

// DefaultWAFConfig returns default WAF configuration
func DefaultWAFConfig() *WAFConfig {
	return &WAFConfig{
		Enabled:           true,
		Mode:              "prevention", // detection, prevention
		LogBlocked:        true,
		BlockByDefault:    false,
		EnableOWASPCRS:    true,
		MaxRequestSize:    10 * 1024 * 1024, // 10MB
		MaxHeaderSize:     8192,              // 8KB
		MaxQueryParams:    100,
		MaxPostParams:     100,
		BlockedResponse:   "Request blocked by Web Application Firewall",
		AllowedMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowedExtensions: []string{},
		BlockedExtensions: []string{"php", "asp", "aspx", "jsp", "pl", "py", "sh", "exe", "bat"},
		GeoBlocking: &GeoBlockingConfig{
			Enabled:         false,
			BlockedCountries: []string{},
			AllowedCountries: []string{},
			DefaultAction:   "allow",
		},
		RateLimiting: &WAFRateLimitConfig{
			Enabled:        false,
			RequestsPerMin: 100,
			BurstSize:      20,
			BanDuration:    10 * time.Minute,
		},
	}
}

// Matcher implementations (simplified)

type SignatureMatcher struct {
	rule   *WAFRule
	target WAFTarget
}

func NewSignatureMatcher(rule *WAFRule, target WAFTarget) *SignatureMatcher {
	return &SignatureMatcher{rule: rule, target: target}
}

func (sm *SignatureMatcher) Match(ctx *RequestContext) (bool, *MatchResult) {
	value := sm.getTargetValue(ctx)
	if sm.rule.Regex != nil && sm.rule.Regex.MatchString(value) {
		return true, &MatchResult{
			Matched:    true,
			Rule:       sm.rule,
			Target:     sm.target,
			Value:      value,
			Evidence:   sm.rule.Regex.FindString(value),
			Confidence: 0.9,
		}
	}
	return false, nil
}

func (sm *SignatureMatcher) GetType() string {
	return "signature"
}

func (sm *SignatureMatcher) getTargetValue(ctx *RequestContext) string {
	switch sm.target {
	case WAFTargetURL:
		return ctx.URL.String()
	case WAFTargetQueryString:
		return ctx.URL.RawQuery
	case WAFTargetBody:
		return string(ctx.Body)
	case WAFTargetUserAgent:
		return ctx.UserAgent
	case WAFTargetReferer:
		return ctx.Referer
	default:
		return ""
	}
}

type RegexMatcher struct {
	*SignatureMatcher
}

func NewRegexMatcher(rule *WAFRule, target WAFTarget) *RegexMatcher {
	return &RegexMatcher{NewSignatureMatcher(rule, target)}
}

func (rm *RegexMatcher) GetType() string {
	return "regex"
}

type BehavioralMatcher struct {
	rule   *WAFRule
	target WAFTarget
}

func NewBehavioralMatcher(rule *WAFRule, target WAFTarget) *BehavioralMatcher {
	return &BehavioralMatcher{rule: rule, target: target}
}

func (bm *BehavioralMatcher) Match(ctx *RequestContext) (bool, *MatchResult) {
	// Simplified behavioral analysis
	// In real implementation, this would analyze request patterns, timing, etc.
	return false, nil
}

func (bm *BehavioralMatcher) GetType() string {
	return "behavioral"
}