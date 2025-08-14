package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// EnterpriseIntegrationManager manages all enterprise integrations
type EnterpriseIntegrationManager struct {
	servicenow *ServiceNowIntegration
	jira       *JiraIntegration
	slack      *SlackIntegration
	datadog    *DatadogIntegration
	newrelic   *NewRelicIntegration
	config     *EnterpriseConfig
}

// EnterpriseConfig holds configuration for enterprise integrations
type EnterpriseConfig struct {
	ServiceNow *ServiceNowConfig `json:"servicenow,omitempty"`
	Jira       *JiraConfig       `json:"jira,omitempty"`
	Slack      *SlackConfig      `json:"slack,omitempty"`
	Datadog    *DatadogConfig    `json:"datadog,omitempty"`
	NewRelic   *NewRelicConfig   `json:"newrelic,omitempty"`
}

// ServiceNowConfig configures ServiceNow integration
type ServiceNowConfig struct {
	Instance string `json:"instance"`
	Username string `json:"username"`
	Password string `json:"password"`
	Enabled  bool   `json:"enabled"`
}

// JiraConfig configures Jira integration
type JiraConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	APIToken string `json:"api_token"`
	Project  string `json:"project"`
	Enabled  bool   `json:"enabled"`
}

// SlackConfig configures Slack integration
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
	Enabled    bool   `json:"enabled"`
}

// DatadogConfig configures Datadog integration
type DatadogConfig struct {
	APIKey  string `json:"api_key"`
	AppKey  string `json:"app_key"`
	Site    string `json:"site"`
	Enabled bool   `json:"enabled"`
}

// NewRelicConfig configures New Relic integration
type NewRelicConfig struct {
	APIKey    string `json:"api_key"`
	AccountID string `json:"account_id"`
	Enabled   bool   `json:"enabled"`
}

// NewEnterpriseIntegrationManager creates a new enterprise integration manager
func NewEnterpriseIntegrationManager(config *EnterpriseConfig) *EnterpriseIntegrationManager {
	if config == nil {
		config = &EnterpriseConfig{}
	}

	manager := &EnterpriseIntegrationManager{
		config: config,
	}

	// Initialize integrations based on configuration
	if config.ServiceNow != nil && config.ServiceNow.Enabled {
		manager.servicenow = NewServiceNowIntegration(config.ServiceNow)
	}

	if config.Jira != nil && config.Jira.Enabled {
		manager.jira = NewJiraIntegration(config.Jira)
	}

	if config.Slack != nil && config.Slack.Enabled {
		manager.slack = NewSlackIntegration(config.Slack)
	}

	if config.Datadog != nil && config.Datadog.Enabled {
		manager.datadog = NewDatadogIntegration(config.Datadog)
	}

	if config.NewRelic != nil && config.NewRelic.Enabled {
		manager.newrelic = NewNewRelicIntegration(config.NewRelic)
	}

	return manager
}

// ServiceNowIntegration handles ServiceNow integration
type ServiceNowIntegration struct {
	config *ServiceNowConfig
	client *http.Client
}

// NewServiceNowIntegration creates a new ServiceNow integration
func NewServiceNowIntegration(config *ServiceNowConfig) *ServiceNowIntegration {
	return &ServiceNowIntegration{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateIncident creates an incident in ServiceNow
func (sn *ServiceNowIntegration) CreateIncident(title, description, severity string) error {
	incident := map[string]interface{}{
		"short_description": title,
		"description":       description,
		"urgency":           severity,
		"impact":            severity,
		"category":          "Software",
		"subcategory":       "Application",
	}

	data, err := json.Marshal(incident)
	if err != nil {
		return fmt.Errorf("failed to marshal incident data: %w", err)
	}

	url := fmt.Sprintf("https://%s.service-now.com/api/now/table/incident", sn.config.Instance)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(sn.config.Username, sn.config.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := sn.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("ServiceNow API returned status %d", resp.StatusCode)
	}

	log.Info().Str("title", title).Msg("ServiceNow incident created")
	return nil
}

// JiraIntegration handles Jira integration
type JiraIntegration struct {
	config *JiraConfig
	client *http.Client
}

// NewJiraIntegration creates a new Jira integration
func NewJiraIntegration(config *JiraConfig) *JiraIntegration {
	return &JiraIntegration{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateIssue creates an issue in Jira
func (j *JiraIntegration) CreateIssue(summary, description, issueType string) error {
	issue := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.config.Project,
			},
			"summary":     summary,
			"description": description,
			"issuetype": map[string]string{
				"name": issueType,
			},
		},
	}

	data, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("failed to marshal issue data: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/2/issue", j.config.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(j.config.Username, j.config.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Jira API returned status %d", resp.StatusCode)
	}

	log.Info().Str("summary", summary).Msg("Jira issue created")
	return nil
}

// SlackIntegration handles Slack integration
type SlackIntegration struct {
	config *SlackConfig
	client *http.Client
}

// NewSlackIntegration creates a new Slack integration
func NewSlackIntegration(config *SlackConfig) *SlackIntegration {
	return &SlackIntegration{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendMessage sends a message to Slack
func (s *SlackIntegration) SendMessage(message, level string) error {
	payload := map[string]interface{}{
		"channel":    s.config.Channel,
		"username":   s.config.Username,
		"text":       message,
		"icon_emoji": s.getEmojiForLevel(level),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message data: %w", err)
	}

	req, err := http.NewRequest("POST", s.config.WebhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status %d", resp.StatusCode)
	}

	log.Info().Str("message", message).Msg("Slack message sent")
	return nil
}

func (s *SlackIntegration) getEmojiForLevel(level string) string {
	switch level {
	case "error", "critical":
		return ":red_circle:"
	case "warning":
		return ":warning:"
	case "info":
		return ":information_source:"
	default:
		return ":white_circle:"
	}
}

// DatadogIntegration handles Datadog integration
type DatadogIntegration struct {
	config *DatadogConfig
	client *http.Client
}

// NewDatadogIntegration creates a new Datadog integration
func NewDatadogIntegration(config *DatadogConfig) *DatadogIntegration {
	return &DatadogIntegration{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendMetric sends a metric to Datadog
func (d *DatadogIntegration) SendMetric(name string, value float64, tags []string) error {
	metric := map[string]interface{}{
		"series": []map[string]interface{}{
			{
				"metric": name,
				"points": [][]interface{}{
					{time.Now().Unix(), value},
				},
				"tags": tags,
			},
		},
	}

	data, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric data: %w", err)
	}

	url := fmt.Sprintf("https://api.%s/api/v1/series", d.config.Site)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", d.config.APIKey)
	req.Header.Set("DD-APPLICATION-KEY", d.config.AppKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Datadog API returned status %d", resp.StatusCode)
	}

	log.Debug().Str("metric", name).Float64("value", value).Msg("Datadog metric sent")
	return nil
}

// NewRelicIntegration handles New Relic integration
type NewRelicIntegration struct {
	config *DatadogConfig
	client *http.Client
}

// NewNewRelicIntegration creates a new New Relic integration
func NewNewRelicIntegration(config *NewRelicConfig) *NewRelicIntegration {
	return &NewRelicIntegration{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendEvent sends an event to New Relic
func (nr *NewRelicIntegration) SendEvent(eventType string, attributes map[string]interface{}) error {
	event := []map[string]interface{}{
		{
			"eventType": eventType,
		},
	}

	// Add attributes to the event
	for k, v := range attributes {
		event[0][k] = v
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	url := "https://insights-collector.newrelic.com/v1/accounts/YOUR_ACCOUNT_ID/events"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Insert-Key", "YOUR_INSERT_KEY")

	resp, err := nr.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("New Relic API returned status %d", resp.StatusCode)
	}

	log.Debug().Str("eventType", eventType).Msg("New Relic event sent")
	return nil
}

// Enterprise Integration Methods

// NotifyIncident notifies all configured systems about an incident
func (eim *EnterpriseIntegrationManager) NotifyIncident(title, description, severity string) {
	ctx := context.Background()

	// ServiceNow
	if eim.servicenow != nil {
		go func() {
			if err := eim.servicenow.CreateIncident(title, description, severity); err != nil {
				log.Error().Err(err).Msg("Failed to create ServiceNow incident")
			}
		}()
	}

	// Jira
	if eim.jira != nil {
		go func() {
			if err := eim.jira.CreateIssue(title, description, "Bug"); err != nil {
				log.Error().Err(err).Msg("Failed to create Jira issue")
			}
		}()
	}

	// Slack
	if eim.slack != nil {
		go func() {
			message := fmt.Sprintf("🚨 *%s*\n%s\nSeverity: %s", title, description, severity)
			if err := eim.slack.SendMessage(message, severity); err != nil {
				log.Error().Err(err).Msg("Failed to send Slack message")
			}
		}()
	}

	_ = ctx // Use context for cancellation if needed
}

// SendMetrics sends metrics to all configured monitoring systems
func (eim *EnterpriseIntegrationManager) SendMetrics(metrics map[string]float64, tags []string) {
	// Datadog
	if eim.datadog != nil {
		go func() {
			for name, value := range metrics {
				if err := eim.datadog.SendMetric(name, value, tags); err != nil {
					log.Error().Err(err).Str("metric", name).Msg("Failed to send Datadog metric")
				}
			}
		}()
	}

	// New Relic
	if eim.newrelic != nil {
		go func() {
			attributes := make(map[string]interface{})
			for k, v := range metrics {
				attributes[k] = v
			}
			for i, tag := range tags {
				attributes[fmt.Sprintf("tag_%d", i)] = tag
			}

			if err := eim.newrelic.SendEvent("OllamaMaxMetrics", attributes); err != nil {
				log.Error().Err(err).Msg("Failed to send New Relic event")
			}
		}()
	}
}

// GetIntegrationStatus returns the status of all integrations
func (eim *EnterpriseIntegrationManager) GetIntegrationStatus() map[string]bool {
	status := make(map[string]bool)

	status["servicenow"] = eim.servicenow != nil
	status["jira"] = eim.jira != nil
	status["slack"] = eim.slack != nil
	status["datadog"] = eim.datadog != nil
	status["newrelic"] = eim.newrelic != nil

	return status
}
