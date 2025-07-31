package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// NotificationSystem manages alert notifications and integrations
type NotificationSystem struct {
	config    *NotificationConfig
	providers map[string]NotificationProvider
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NotificationConfig configures the notification system
type NotificationConfig struct {
	Enabled bool `json:"enabled"`

	// Slack configuration
	SlackWebhookURL string `json:"slack_webhook_url"`
	SlackChannel    string `json:"slack_channel"`

	// Email configuration
	SMTPHost     string   `json:"smtp_host"`
	SMTPPort     int      `json:"smtp_port"`
	SMTPUsername string   `json:"smtp_username"`
	SMTPPassword string   `json:"smtp_password"`
	EmailFrom    string   `json:"email_from"`
	EmailTo      []string `json:"email_to"`

	// Webhook configuration
	WebhookURLs []string `json:"webhook_urls"`

	// Rate limiting
	RateLimitWindow  time.Duration `json:"rate_limit_window"`
	MaxNotifications int           `json:"max_notifications"`
}

// NotificationProvider interface for different notification channels
type NotificationProvider interface {
	SendNotification(ctx context.Context, notification *Notification) error
	GetName() string
	IsEnabled() bool
}

// Notification represents an alert notification
type Notification struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Severity    string                 `json:"severity"`
	Component   string                 `json:"component"`
	NodeID      string                 `json:"node_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SlackProvider implements Slack notifications
type SlackProvider struct {
	webhookURL string
	channel    string
	enabled    bool
}

// EmailProvider implements email notifications
type EmailProvider struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	from         string
	to           []string
	enabled      bool
}

// WebhookProvider implements webhook notifications
type WebhookProvider struct {
	urls    []string
	enabled bool
}

// NewNotificationSystem creates a new notification system
func NewNotificationSystem(config *NotificationConfig) *NotificationSystem {
	if config == nil {
		config = &NotificationConfig{
			Enabled:          false,
			RateLimitWindow:  time.Minute,
			MaxNotifications: 10,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	ns := &NotificationSystem{
		config:    config,
		providers: make(map[string]NotificationProvider),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize providers
	if config.Enabled {
		ns.initializeProviders()
	}

	return ns
}

// initializeProviders sets up notification providers
func (ns *NotificationSystem) initializeProviders() {
	// Slack provider
	if ns.config.SlackWebhookURL != "" {
		slack := &SlackProvider{
			webhookURL: ns.config.SlackWebhookURL,
			channel:    ns.config.SlackChannel,
			enabled:    true,
		}
		ns.providers["slack"] = slack
		log.Info().Msg("Slack notification provider initialized")
	}

	// Email provider
	if ns.config.SMTPHost != "" && len(ns.config.EmailTo) > 0 {
		email := &EmailProvider{
			smtpHost:     ns.config.SMTPHost,
			smtpPort:     ns.config.SMTPPort,
			smtpUsername: ns.config.SMTPUsername,
			smtpPassword: ns.config.SMTPPassword,
			from:         ns.config.EmailFrom,
			to:           ns.config.EmailTo,
			enabled:      true,
		}
		ns.providers["email"] = email
		log.Info().Msg("Email notification provider initialized")
	}

	// Webhook provider
	if len(ns.config.WebhookURLs) > 0 {
		webhook := &WebhookProvider{
			urls:    ns.config.WebhookURLs,
			enabled: true,
		}
		ns.providers["webhook"] = webhook
		log.Info().Msg("Webhook notification provider initialized")
	}
}

// SendNotification sends a notification through all enabled providers
func (ns *NotificationSystem) SendNotification(notification *Notification) error {
	if !ns.config.Enabled {
		return nil
	}

	ns.mu.RLock()
	defer ns.mu.RUnlock()

	var errors []error

	for name, provider := range ns.providers {
		if !provider.IsEnabled() {
			continue
		}

		go func(providerName string, p NotificationProvider) {
			ctx, cancel := context.WithTimeout(ns.ctx, 30*time.Second)
			defer cancel()

			if err := p.SendNotification(ctx, notification); err != nil {
				log.Error().
					Err(err).
					Str("provider", providerName).
					Str("notification_id", notification.ID).
					Msg("Failed to send notification")
			} else {
				log.Info().
					Str("provider", providerName).
					Str("notification_id", notification.ID).
					Msg("Notification sent successfully")
			}
		}(name, provider)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send notifications: %v", errors)
	}

	return nil
}

// SendHealthAlert sends a health-related alert
func (ns *NotificationSystem) SendHealthAlert(component, status, message string, metadata map[string]interface{}) error {
	notification := &Notification{
		ID:        fmt.Sprintf("health-%s-%d", component, time.Now().Unix()),
		Title:     fmt.Sprintf("Health Alert: %s", component),
		Message:   message,
		Severity:  "warning",
		Component: component,
		Timestamp: time.Now(),
		Labels: map[string]string{
			"component": component,
			"status":    status,
			"type":      "health",
		},
		Metadata: metadata,
	}

	return ns.SendNotification(notification)
}

// SendMetricAlert sends a metric-based alert
func (ns *NotificationSystem) SendMetricAlert(metric, threshold, currentValue string, metadata map[string]interface{}) error {
	notification := &Notification{
		ID:        fmt.Sprintf("metric-%s-%d", metric, time.Now().Unix()),
		Title:     fmt.Sprintf("Metric Alert: %s", metric),
		Message:   fmt.Sprintf("Metric %s exceeded threshold %s (current: %s)", metric, threshold, currentValue),
		Severity:  "warning",
		Component: "metrics",
		Timestamp: time.Now(),
		Labels: map[string]string{
			"metric":    metric,
			"threshold": threshold,
			"value":     currentValue,
			"type":      "metric",
		},
		Metadata: metadata,
	}

	return ns.SendNotification(notification)
}

// Shutdown gracefully shuts down the notification system
func (ns *NotificationSystem) Shutdown() error {
	ns.cancel()
	log.Info().Msg("Notification system stopped")
	return nil
}

// Slack Provider Implementation
func (sp *SlackProvider) SendNotification(ctx context.Context, notification *Notification) error {
	payload := map[string]interface{}{
		"channel": sp.channel,
		"text":    notification.Title,
		"attachments": []map[string]interface{}{
			{
				"color": sp.getSeverityColor(notification.Severity),
				"fields": []map[string]interface{}{
					{
						"title": "Message",
						"value": notification.Message,
						"short": false,
					},
					{
						"title": "Component",
						"value": notification.Component,
						"short": true,
					},
					{
						"title": "Severity",
						"value": notification.Severity,
						"short": true,
					},
					{
						"title": "Timestamp",
						"value": notification.Timestamp.Format(time.RFC3339),
						"short": true,
					},
				},
			},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sp.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status %d", resp.StatusCode)
	}

	return nil
}

func (sp *SlackProvider) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "danger"
	case "warning":
		return "warning"
	case "info":
		return "good"
	default:
		return "#439FE0"
	}
}

func (sp *SlackProvider) GetName() string {
	return "slack"
}

func (sp *SlackProvider) IsEnabled() bool {
	return sp.enabled
}

// Email Provider Implementation
func (ep *EmailProvider) SendNotification(ctx context.Context, notification *Notification) error {
	subject := fmt.Sprintf("[%s] %s", notification.Severity, notification.Title)
	body := fmt.Sprintf(`
Alert: %s
Message: %s
Component: %s
Severity: %s
Node ID: %s
Timestamp: %s

Labels:
%s

Metadata:
%s
`,
		notification.Title,
		notification.Message,
		notification.Component,
		notification.Severity,
		notification.NodeID,
		notification.Timestamp.Format(time.RFC3339),
		ep.formatLabels(notification.Labels),
		ep.formatMetadata(notification.Metadata),
	)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		ep.from, ep.to[0], subject, body)

	auth := smtp.PlainAuth("", ep.smtpUsername, ep.smtpPassword, ep.smtpHost)
	addr := fmt.Sprintf("%s:%d", ep.smtpHost, ep.smtpPort)

	return smtp.SendMail(addr, auth, ep.from, ep.to, []byte(msg))
}

func (ep *EmailProvider) formatLabels(labels map[string]string) string {
	var result string
	for k, v := range labels {
		result += fmt.Sprintf("  %s: %s\n", k, v)
	}
	return result
}

func (ep *EmailProvider) formatMetadata(metadata map[string]interface{}) string {
	var result string
	for k, v := range metadata {
		result += fmt.Sprintf("  %s: %v\n", k, v)
	}
	return result
}

func (ep *EmailProvider) GetName() string {
	return "email"
}

func (ep *EmailProvider) IsEnabled() bool {
	return ep.enabled
}

// Webhook Provider Implementation
func (wp *WebhookProvider) SendNotification(ctx context.Context, notification *Notification) error {
	jsonPayload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	for _, url := range wp.urls {
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send webhook notification: %w", err)
		}
		resp.Body.Close()
	}

	return nil
}

func (wp *WebhookProvider) GetName() string {
	return "webhook"
}

func (wp *WebhookProvider) IsEnabled() bool {
	return wp.enabled
}

// DefaultNotificationConfig returns a default notification configuration
func DefaultNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		Enabled:          false,
		SlackChannel:     "#ollama-alerts",
		SMTPPort:         587,
		EmailFrom:        "alerts@ollama-distributed.com",
		RateLimitWindow:  time.Minute,
		MaxNotifications: 10,
	}
}
