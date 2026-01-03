package discord

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/maxsupermanhd/FactoCord-3.0/v3/support"
)

// VerificationData stores the link between Discord user IDs and Factorio usernames
type VerificationData struct {
	// Maps Discord user ID to Factorio username
	DiscordToFactorio map[string]string `json:"discord_to_factorio"`
	// Maps Factorio username to Discord user ID
	FactorioToDiscord map[string]string `json:"factorio_to_discord"`
}

// PendingVerification stores pending verification requests
type PendingVerification struct {
	DiscordUserID    string
	FactorioUsername string
	Code             string
	ExpiresAt        time.Time
}

var (
	verificationData     VerificationData
	pendingVerifications map[string]*PendingVerification // key: Discord user ID
	verificationMutex    sync.RWMutex
	pendingMutex         sync.RWMutex
)

func init() {
	verificationData = VerificationData{
		DiscordToFactorio: make(map[string]string),
		FactorioToDiscord: make(map[string]string),
	}
	pendingVerifications = make(map[string]*PendingVerification)
}

// LoadVerificationData loads verification data from file
func LoadVerificationData() error {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()

	path := support.Config.VerificationDataPath
	if path == "" {
		path = "./verification.json"
	}

	if !support.FileExists(path) {
		return nil // No data file yet, that's okay
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read verification data: %w", err)
	}

	err = json.Unmarshal(data, &verificationData)
	if err != nil {
		return fmt.Errorf("failed to parse verification data: %w", err)
	}

	return nil
}

// SaveVerificationData saves verification data to file
func SaveVerificationData() error {
	verificationMutex.RLock()
	defer verificationMutex.RUnlock()

	path := support.Config.VerificationDataPath
	if path == "" {
		path = "./verification.json"
	}

	data, err := json.MarshalIndent(verificationData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal verification data: %w", err)
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write verification data: %w", err)
	}

	return nil
}

// GenerateVerificationCode generates a random 6-digit code
func GenerateVerificationCode() (string, error) {
	// Generate 6 random digits
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}

// CreateVerificationRequest creates a new pending verification
func CreateVerificationRequest(discordUserID, factorioUsername string) (string, error) {
	code, err := GenerateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}

	pendingMutex.Lock()
	defer pendingMutex.Unlock()

	// Remove any existing pending verification for this user
	delete(pendingVerifications, discordUserID)

	// Create new pending verification (expires in 5 minutes)
	pendingVerifications[discordUserID] = &PendingVerification{
		DiscordUserID:    discordUserID,
		FactorioUsername: factorioUsername,
		Code:             code,
		ExpiresAt:        time.Now().Add(5 * time.Minute),
	}

	return code, nil
}

// VerifyCode verifies a code and links the accounts if correct
func VerifyCode(discordUserID, code string) (bool, string, error) {
	pendingMutex.Lock()
	pending, exists := pendingVerifications[discordUserID]
	if !exists {
		pendingMutex.Unlock()
		return false, "", fmt.Errorf("no pending verification found")
	}

	if time.Now().After(pending.ExpiresAt) {
		delete(pendingVerifications, discordUserID)
		pendingMutex.Unlock()
		return false, "", fmt.Errorf("verification code has expired")
	}

	if pending.Code != code {
		pendingMutex.Unlock()
		return false, "", nil // Wrong code but not an error
	}

	factorioUsername := pending.FactorioUsername
	delete(pendingVerifications, discordUserID)
	pendingMutex.Unlock()

	// Link the accounts
	verificationMutex.Lock()
	// Remove any old links for this Discord user
	if oldFactorio, exists := verificationData.DiscordToFactorio[discordUserID]; exists {
		delete(verificationData.FactorioToDiscord, oldFactorio)
	}
	// Remove any old links for this Factorio user
	if oldDiscord, exists := verificationData.FactorioToDiscord[factorioUsername]; exists {
		delete(verificationData.DiscordToFactorio, oldDiscord)
	}

	verificationData.DiscordToFactorio[discordUserID] = factorioUsername
	verificationData.FactorioToDiscord[factorioUsername] = discordUserID
	verificationMutex.Unlock()

	// Save to file
	err := SaveVerificationData()
	if err != nil {
		return true, factorioUsername, fmt.Errorf("linked but failed to save: %w", err)
	}

	return true, factorioUsername, nil
}

// GetFactorioUsername gets the linked Factorio username for a Discord user
func GetFactorioUsername(discordUserID string) (string, bool) {
	verificationMutex.RLock()
	defer verificationMutex.RUnlock()
	username, exists := verificationData.DiscordToFactorio[discordUserID]
	return username, exists
}

// GetDiscordUserID gets the linked Discord user ID for a Factorio username
func GetDiscordUserID(factorioUsername string) (string, bool) {
	verificationMutex.RLock()
	defer verificationMutex.RUnlock()
	userID, exists := verificationData.FactorioToDiscord[factorioUsername]
	return userID, exists
}

// IsUserVerified checks if a Discord user is verified
func IsUserVerified(discordUserID string) bool {
	_, exists := GetFactorioUsername(discordUserID)
	return exists
}

// UnlinkUser removes the verification link for a Discord user
func UnlinkUser(discordUserID string) error {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()

	factorioUsername, exists := verificationData.DiscordToFactorio[discordUserID]
	if !exists {
		return fmt.Errorf("user is not verified")
	}

	delete(verificationData.DiscordToFactorio, discordUserID)
	delete(verificationData.FactorioToDiscord, factorioUsername)

	return SaveVerificationData()
}

// GetPendingVerification gets the pending verification for a Discord user
func GetPendingVerification(discordUserID string) (*PendingVerification, bool) {
	pendingMutex.RLock()
	defer pendingMutex.RUnlock()
	pending, exists := pendingVerifications[discordUserID]
	if !exists || time.Now().After(pending.ExpiresAt) {
		return nil, false
	}
	return pending, true
}

// CleanupExpiredVerifications removes expired pending verifications
func CleanupExpiredVerifications() {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()

	now := time.Now()
	for key, pending := range pendingVerifications {
		if now.After(pending.ExpiresAt) {
			delete(pendingVerifications, key)
		}
	}
}

// StartVerificationCleanup starts a background goroutine to clean up expired verifications
func StartVerificationCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			CleanupExpiredVerifications()
		}
	}()
}
