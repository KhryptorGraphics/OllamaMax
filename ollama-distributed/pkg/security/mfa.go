package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	_ "crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// MFAManager handles Multi-Factor Authentication
type MFAManager struct {
	config           *MFAConfig
	totpManager      *TOTPManager
	smsManager       *SMSManager
	emailManager     *EmailManager
	hardwareManager  *HardwareKeyManager
	backupCodes      map[string][]string // userID -> backup codes
	userMFASettings  map[string]*UserMFASettings
	mu               sync.RWMutex
}

// MFAConfig configures MFA settings
type MFAConfig struct {
	EnabledMethods    []string      `json:"enabled_methods"` // totp, sms, email, hardware_key, backup_codes
	TOTPIssuer        string        `json:"totp_issuer"`
	TOTPDigits        int           `json:"totp_digits"`
	TOTPPeriod        int           `json:"totp_period"`
	SMSProvider       string        `json:"sms_provider"`
	SMSConfig         interface{}   `json:"sms_config"`
	EmailProvider     string        `json:"email_provider"`
	EmailConfig       interface{}   `json:"email_config"`
	BackupCodesCount  int           `json:"backup_codes_count"`
	RequireMFA        bool          `json:"require_mfa"`
	GracePeriod       time.Duration `json:"grace_period"`
	MaxAttempts       int           `json:"max_attempts"`
	LockoutDuration   time.Duration `json:"lockout_duration"`
}

// UserMFASettings stores MFA settings for a user
type UserMFASettings struct {
	UserID             string                    `json:"user_id"`
	EnabledMethods     []string                  `json:"enabled_methods"`
	TOTPSecret         string                    `json:"totp_secret,omitempty"`
	TOTPVerified       bool                      `json:"totp_verified"`
	PhoneNumber        string                    `json:"phone_number,omitempty"`
	PhoneVerified      bool                      `json:"phone_verified"`
	Email              string                    `json:"email,omitempty"`
	EmailVerified      bool                      `json:"email_verified"`
	BackupCodes        []string                  `json:"backup_codes,omitempty"`
	HardwareKeys       []HardwareKeyInfo         `json:"hardware_keys,omitempty"`
	Attempts           map[string]int            `json:"attempts"` // method -> attempt count
	LastAttempt        map[string]time.Time      `json:"last_attempt"`
	LockedUntil        map[string]time.Time      `json:"locked_until"`
	TrustedDevices     []TrustedDevice           `json:"trusted_devices"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}

// TOTPManager handles TOTP (Time-based One-Time Password)
type TOTPManager struct {
	issuer string
	digits int
	period int
}

// SMSManager handles SMS-based MFA
type SMSManager struct {
	provider string
	config   interface{}
	client   interface{}
}

// EmailManager handles Email-based MFA
type EmailManager struct {
	provider string
	config   interface{}
	client   interface{}
}

// HardwareKeyManager handles hardware key MFA (FIDO2/WebAuthn)
type HardwareKeyManager struct {
	rpID     string
	rpName   string
	rpOrigin string
}

// HardwareKeyInfo stores hardware key information
type HardwareKeyInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CredentialID string    `json:"credential_id"`
	PublicKey    []byte    `json:"public_key"`
	Counter      uint32    `json:"counter"`
	CreatedAt    time.Time `json:"created_at"`
}

// TrustedDevice represents a trusted device
type TrustedDevice struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	UserAgent   string    `json:"user_agent"`
	IPAddress   string    `json:"ip_address"`
	Fingerprint string    `json:"fingerprint"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// MFAChallenge represents an MFA challenge
type MFAChallenge struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	Method      string            `json:"method"`
	Challenge   string            `json:"challenge,omitempty"`
	ExpiresAt   time.Time         `json:"expires_at"`
	Verified    bool              `json:"verified"`
	Attempts    int               `json:"attempts"`
	Metadata    map[string]string `json:"metadata"`
}

// MFAVerificationResult represents MFA verification result
type MFAVerificationResult struct {
	Success       bool              `json:"success"`
	Method        string            `json:"method"`
	UserID        string            `json:"user_id"`
	TrustedDevice bool              `json:"trusted_device"`
	Metadata      map[string]string `json:"metadata"`
	Error         string            `json:"error,omitempty"`
}

// NewMFAManager creates a new MFA manager
func NewMFAManager(methods []string) *MFAManager {
	config := &MFAConfig{
		EnabledMethods:    methods,
		TOTPIssuer:        "OllamaMax",
		TOTPDigits:        6,
		TOTPPeriod:        30,
		BackupCodesCount:  10,
		RequireMFA:        true,
		GracePeriod:       7 * 24 * time.Hour,
		MaxAttempts:       3,
		LockoutDuration:   15 * time.Minute,
	}

	manager := &MFAManager{
		config:          config,
		backupCodes:     make(map[string][]string),
		userMFASettings: make(map[string]*UserMFASettings),
	}

	// Initialize TOTP manager
	if contains(methods, "totp") {
		manager.totpManager = &TOTPManager{
			issuer: config.TOTPIssuer,
			digits: config.TOTPDigits,
			period: config.TOTPPeriod,
		}
	}

	// Initialize SMS manager
	if contains(methods, "sms") {
		manager.smsManager = &SMSManager{
			provider: config.SMSProvider,
			config:   config.SMSConfig,
		}
	}

	// Initialize Email manager
	if contains(methods, "email") {
		manager.emailManager = &EmailManager{
			provider: config.EmailProvider,
			config:   config.EmailConfig,
		}
	}

	// Initialize Hardware Key manager
	if contains(methods, "hardware_key") {
		manager.hardwareManager = &HardwareKeyManager{
			rpID:     "localhost", // Should be configured properly
			rpName:   "OllamaMax",
			rpOrigin: "https://localhost:8080",
		}
	}

	log.Info().
		Strs("methods", methods).
		Msg("MFA manager initialized")

	return manager
}

// SetupMFA sets up MFA for a user
func (mfa *MFAManager) SetupMFA(userID string, method string) (*MFAChallenge, error) {
	mfa.mu.Lock()
	defer mfa.mu.Unlock()

	// Get or create user MFA settings
	settings, exists := mfa.userMFASettings[userID]
	if !exists {
		settings = &UserMFASettings{
			UserID:         userID,
			EnabledMethods: []string{},
			Attempts:       make(map[string]int),
			LastAttempt:    make(map[string]time.Time),
			LockedUntil:    make(map[string]time.Time),
			TrustedDevices: []TrustedDevice{},
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		mfa.userMFASettings[userID] = settings
	}

	switch method {
	case "totp":
		return mfa.setupTOTP(userID, settings)
	case "sms":
		return mfa.setupSMS(userID, settings)
	case "email":
		return mfa.setupEmail(userID, settings)
	case "hardware_key":
		return mfa.setupHardwareKey(userID, settings)
	case "backup_codes":
		return mfa.setupBackupCodes(userID, settings)
	default:
		return nil, fmt.Errorf("unsupported MFA method: %s", method)
	}
}

// setupTOTP sets up TOTP for a user
func (mfa *MFAManager) setupTOTP(userID string, settings *UserMFASettings) (*MFAChallenge, error) {
	if mfa.totpManager == nil {
		return nil, fmt.Errorf("TOTP not enabled")
	}

	// Generate secret
	secret, err := mfa.generateTOTPSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	settings.TOTPSecret = secret
	settings.TOTPVerified = false
	settings.UpdatedAt = time.Now()

	// Generate QR code URL
	qrURL := mfa.generateTOTPURL(userID, secret)

	challenge := &MFAChallenge{
		ID:        generateChallengeID(),
		UserID:    userID,
		Method:    "totp",
		Challenge: qrURL,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Verified:  false,
		Metadata: map[string]string{
			"secret": secret,
		},
	}

	return challenge, nil
}

// setupSMS sets up SMS MFA for a user
func (mfa *MFAManager) setupSMS(userID string, settings *UserMFASettings) (*MFAChallenge, error) {
	if mfa.smsManager == nil {
		return nil, fmt.Errorf("SMS MFA not enabled")
	}

	// Generate verification code
	code := mfa.generateVerificationCode()

	challenge := &MFAChallenge{
		ID:        generateChallengeID(),
		UserID:    userID,
		Method:    "sms",
		Challenge: code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Verified:  false,
		Metadata: map[string]string{
			"code": code,
		},
	}

	// Send SMS (implementation would send actual SMS)
	log.Info().
		Str("user_id", userID).
		Str("method", "sms").
		Msg("SMS verification code generated")

	return challenge, nil
}

// setupEmail sets up Email MFA for a user
func (mfa *MFAManager) setupEmail(userID string, settings *UserMFASettings) (*MFAChallenge, error) {
	if mfa.emailManager == nil {
		return nil, fmt.Errorf("Email MFA not enabled")
	}

	// Generate verification code
	code := mfa.generateVerificationCode()

	challenge := &MFAChallenge{
		ID:        generateChallengeID(),
		UserID:    userID,
		Method:    "email",
		Challenge: code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Verified:  false,
		Metadata: map[string]string{
			"code": code,
		},
	}

	// Send email (implementation would send actual email)
	log.Info().
		Str("user_id", userID).
		Str("method", "email").
		Msg("Email verification code generated")

	return challenge, nil
}

// setupHardwareKey sets up hardware key MFA
func (mfa *MFAManager) setupHardwareKey(userID string, settings *UserMFASettings) (*MFAChallenge, error) {
	if mfa.hardwareManager == nil {
		return nil, fmt.Errorf("Hardware key MFA not enabled")
	}

	// Generate WebAuthn challenge
	challengeBytes := make([]byte, 32)
	rand.Read(challengeBytes)
	challenge := base32.StdEncoding.EncodeToString(challengeBytes)

	mfaChallenge := &MFAChallenge{
		ID:        generateChallengeID(),
		UserID:    userID,
		Method:    "hardware_key",
		Challenge: challenge,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Verified:  false,
		Metadata: map[string]string{
			"challenge": challenge,
			"rp_id":     mfa.hardwareManager.rpID,
		},
	}

	return mfaChallenge, nil
}

// setupBackupCodes generates backup codes for a user
func (mfa *MFAManager) setupBackupCodes(userID string, settings *UserMFASettings) (*MFAChallenge, error) {
	codes := make([]string, mfa.config.BackupCodesCount)
	for i := range codes {
		codes[i] = mfa.generateBackupCode()
	}

	settings.BackupCodes = codes
	settings.UpdatedAt = time.Now()

	// Store backup codes
	mfa.backupCodes[userID] = codes

	challenge := &MFAChallenge{
		ID:        generateChallengeID(),
		UserID:    userID,
		Method:    "backup_codes",
		ExpiresAt: time.Now().Add(time.Hour), // Longer expiry for backup codes
		Verified:  true, // Backup codes are immediately available
		Metadata: map[string]string{
			"codes": strings.Join(codes, ","),
		},
	}

	return challenge, nil
}

// VerifyToken verifies an MFA token
func (mfa *MFAManager) VerifyToken(userID, method, token string) (bool, error) {
	mfa.mu.Lock()
	defer mfa.mu.Unlock()

	settings, exists := mfa.userMFASettings[userID]
	if !exists {
		return false, fmt.Errorf("user MFA not set up")
	}

	// Check if method is locked out
	if lockedUntil, exists := settings.LockedUntil[method]; exists {
		if time.Now().Before(lockedUntil) {
			return false, fmt.Errorf("method locked due to too many attempts")
		}
		delete(settings.LockedUntil, method)
	}

	// Increment attempt count
	settings.Attempts[method]++
	settings.LastAttempt[method] = time.Now()

	var verified bool
	var err error

	switch method {
	case "totp":
		verified, err = mfa.verifyTOTP(settings.TOTPSecret, token)
	case "sms":
		verified, err = mfa.verifySMS(userID, token)
	case "email":
		verified, err = mfa.verifyEmail(userID, token)
	case "hardware_key":
		verified, err = mfa.verifyHardwareKey(userID, token)
	case "backup_codes":
		verified, err = mfa.verifyBackupCode(userID, token)
	default:
		return false, fmt.Errorf("unsupported MFA method: %s", method)
	}

	if verified {
		// Reset attempt count on success
		settings.Attempts[method] = 0
		return true, nil
	}

	// Check for lockout
	if settings.Attempts[method] >= mfa.config.MaxAttempts {
		settings.LockedUntil[method] = time.Now().Add(mfa.config.LockoutDuration)
		log.Warn().
			Str("user_id", userID).
			Str("method", method).
			Int("attempts", settings.Attempts[method]).
			Msg("MFA method locked due to too many failed attempts")
	}

	return false, err
}

// verifyTOTP verifies a TOTP token
func (mfa *MFAManager) verifyTOTP(secret, token string) (bool, error) {
	if mfa.totpManager == nil {
		return false, fmt.Errorf("TOTP not enabled")
	}

	// Parse token
	tokenInt, err := strconv.ParseInt(token, 10, 32)
	if err != nil {
		return false, fmt.Errorf("invalid token format")
	}

	// Get current time window
	now := time.Now().Unix()
	timeWindow := now / int64(mfa.totpManager.period)

	// Check current and adjacent time windows for clock skew tolerance
	for i := int64(-1); i <= 1; i++ {
		expectedToken := mfa.generateTOTPToken(secret, timeWindow+i)
		if expectedToken == int(tokenInt) {
			return true, nil
		}
	}

	return false, nil
}

// verifySMS verifies an SMS token
func (mfa *MFAManager) verifySMS(userID, token string) (bool, error) {
	// Implementation would verify SMS token
	// For now, simplified verification
	return len(token) == 6, nil
}

// verifyEmail verifies an email token
func (mfa *MFAManager) verifyEmail(userID, token string) (bool, error) {
	// Implementation would verify email token
	// For now, simplified verification
	return len(token) == 6, nil
}

// verifyHardwareKey verifies a hardware key response
func (mfa *MFAManager) verifyHardwareKey(userID, token string) (bool, error) {
	// Implementation would verify WebAuthn response
	// This requires complex cryptographic verification
	return len(token) > 0, nil
}

// verifyBackupCode verifies and consumes a backup code
func (mfa *MFAManager) verifyBackupCode(userID, code string) (bool, error) {
	settings, exists := mfa.userMFASettings[userID]
	if !exists {
		return false, fmt.Errorf("user MFA not set up")
	}

	// Find and remove the backup code
	for i, backupCode := range settings.BackupCodes {
		if backupCode == code {
			// Remove the used backup code
			settings.BackupCodes = append(settings.BackupCodes[:i], settings.BackupCodes[i+1:]...)
			settings.UpdatedAt = time.Now()
			return true, nil
		}
	}

	return false, fmt.Errorf("invalid backup code")
}

// GetUserMFASettings returns MFA settings for a user
func (mfa *MFAManager) GetUserMFASettings(userID string) (*UserMFASettings, error) {
	mfa.mu.RLock()
	defer mfa.mu.RUnlock()

	settings, exists := mfa.userMFASettings[userID]
	if !exists {
		return nil, fmt.Errorf("user MFA not set up")
	}

	// Return a copy to avoid race conditions
	settingsCopy := *settings
	return &settingsCopy, nil
}

// DisableMFA disables MFA for a user
func (mfa *MFAManager) DisableMFA(userID, method string) error {
	mfa.mu.Lock()
	defer mfa.mu.Unlock()

	settings, exists := mfa.userMFASettings[userID]
	if !exists {
		return fmt.Errorf("user MFA not set up")
	}

	switch method {
	case "totp":
		settings.TOTPSecret = ""
		settings.TOTPVerified = false
	case "sms":
		settings.PhoneNumber = ""
		settings.PhoneVerified = false
	case "email":
		settings.EmailVerified = false
	case "backup_codes":
		settings.BackupCodes = nil
		delete(mfa.backupCodes, userID)
	}

	// Remove from enabled methods
	enabledMethods := make([]string, 0)
	for _, m := range settings.EnabledMethods {
		if m != method {
			enabledMethods = append(enabledMethods, m)
		}
	}
	settings.EnabledMethods = enabledMethods
	settings.UpdatedAt = time.Now()

	log.Info().
		Str("user_id", userID).
		Str("method", method).
		Msg("MFA method disabled")

	return nil
}

// Helper functions

func (mfa *MFAManager) generateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(bytes), nil
}

func (mfa *MFAManager) generateTOTPURL(userID, secret string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&digits=%d&period=%d",
		mfa.totpManager.issuer,
		userID,
		secret,
		mfa.totpManager.issuer,
		mfa.totpManager.digits,
		mfa.totpManager.period,
	)
}

func (mfa *MFAManager) generateTOTPToken(secret string, timeWindow int64) int {
	// Decode secret
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return 0
	}

	// Convert time window to bytes
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeWindow))

	// Generate HMAC-SHA1 hash
	mac := hmac.New(sha1.New, key)
	mac.Write(timeBytes)
	hash := mac.Sum(nil)

	// Extract dynamic binary code
	offset := hash[19] & 0x0f
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate token with specified digits
	token := int(truncated) % int(math.Pow10(mfa.totpManager.digits))
	
	return token
}

func (mfa *MFAManager) generateVerificationCode() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	code := int(binary.BigEndian.Uint32(append([]byte{0}, bytes...))) % 1000000
	return fmt.Sprintf("%06d", code)
}

func (mfa *MFAManager) generateBackupCode() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func generateChallengeID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetAvailableMethods returns available MFA methods
func (mfa *MFAManager) GetAvailableMethods() []string {
	return mfa.config.EnabledMethods
}

// IsMethodEnabled checks if an MFA method is enabled
func (mfa *MFAManager) IsMethodEnabled(method string) bool {
	return contains(mfa.config.EnabledMethods, method)
}

// RequiresMFA checks if MFA is required for the system
func (mfa *MFAManager) RequiresMFA() bool {
	return mfa.config.RequireMFA
}