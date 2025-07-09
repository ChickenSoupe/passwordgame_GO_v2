package rules

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

// CyberSecurityRules handles all cybersecurity-themed password rules
type CyberSecurityRules struct {
	mutex                 sync.RWMutex
	updateAlertShown      bool
	updateString          string
	adWatched             bool
	raidUnlockString      string
	blackSquareCount      int
	blackboxRuleValidated bool
	imposterIndices       []int
	imposterOriginalChars []byte
	imposterRuleValidated bool
	lastPasswordLength    int
}

var cyberSecRules = &CyberSecurityRules{
	updateString:     "UPDATE-2024",
	raidUnlockString: "RAID-UNLOCKED",
}

// Rule14UpdateAlert validates the update alert rule
func Rule14UpdateAlert(password string) bool {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()

	// Check if the update string is present in the password
	return strings.Contains(password, cyberSecRules.updateString)
}

// Rule22PDFFile validates the PDF file rule
func Rule22PDFFile(password string) bool {
	// Simple validation - check if "pdf file" is present
	return strings.Contains(strings.ToLower(password), "pdf file")
}

// Rule23PasswordLock validates the RAID unlock rule
func Rule23PasswordLock(password string) bool {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()

	// Check if the RAID unlock string is present
	return strings.Contains(password, cyberSecRules.raidUnlockString)
}

// Rule24RansomwareAttack validates the ransomware defense rule
func Rule24RansomwareAttack(password string) bool {
	cyberSecRules.mutex.Lock()
	defer cyberSecRules.mutex.Unlock()

	// If the rule has already been validated for this session, return true
	if cyberSecRules.blackboxRuleValidated {
		return true
	}

	// Count black squares in the password
	blackSquareCount := strings.Count(password, "⬛")
	cyberSecRules.blackSquareCount = blackSquareCount

	// Rule is satisfied if there are no black squares (user deleted them all)
	if blackSquareCount == 0 {
		// Mark the rule as validated for this session
		cyberSecRules.blackboxRuleValidated = true
		return true
	}

	return false
}

// Rule25InsiderThreat validates the insider threat rule
func Rule25InsiderThreat(password string) bool {
	cyberSecRules.mutex.Lock()
	defer cyberSecRules.mutex.Unlock()

	// Check if the rule has already been validated for this session
	if cyberSecRules.imposterRuleValidated {
		return true
	}

	// If password length changed and we haven't generated indices yet, generate them
	if len(password) != cyberSecRules.lastPasswordLength && len(cyberSecRules.imposterIndices) == 0 {
		cyberSecRules.generateImposterIndices(password)
		cyberSecRules.lastPasswordLength = len(password)
	}

	// Check if all imposter characters have been removed
	if len(password) < 3 || len(cyberSecRules.imposterIndices) == 0 {
		return true // Rule satisfied if password too short or no imposters
	}

	// Check if the imposter characters have been removed
	allRemoved := true
	for i, idx := range cyberSecRules.imposterIndices {
		// If the index is out of bounds or the character at that position has changed
		if idx >= len(password) || (idx < len(password) && password[idx] != cyberSecRules.imposterOriginalChars[i]) {
			continue // This imposter character has been removed or modified
		} else {
			allRemoved = false
			break
		}
	}

	// If all imposter characters have been removed, mark the rule as validated
	if allRemoved {
		cyberSecRules.imposterRuleValidated = true
		return true
	}

	return false
}

// generateImposterIndices creates random indices for imposter characters
func (csr *CyberSecurityRules) generateImposterIndices(password string) {
	if len(password) < 3 {
		csr.imposterIndices = []int{}
		csr.imposterOriginalChars = []byte{}
		return
	}

	rand.Seed(time.Now().UnixNano())
	indices := make(map[int]bool)

	// Generate 3 unique random indices
	for len(indices) < 3 && len(indices) < len(password) {
		idx := rand.Intn(len(password))
		// Avoid spaces and already selected indices
		if password[idx] != ' ' && !indices[idx] {
			indices[idx] = true
		}
	}

	// Convert map to slice
	csr.imposterIndices = make([]int, 0, len(indices))
	csr.imposterOriginalChars = make([]byte, 0, len(indices))

	for idx := range indices {
		csr.imposterIndices = append(csr.imposterIndices, idx)
		// Store the original character at this position
		csr.imposterOriginalChars = append(csr.imposterOriginalChars, password[idx])
	}
}

// GetUpdateString returns the current update string for Rule 14
func GetUpdateString() string {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()
	return cyberSecRules.updateString
}

// SetUpdateAlertShown marks the update alert as shown
func SetUpdateAlertShown(shown bool) {
	cyberSecRules.mutex.Lock()
	defer cyberSecRules.mutex.Unlock()
	cyberSecRules.updateAlertShown = shown
}

// IsUpdateAlertShown returns whether the update alert has been shown
func IsUpdateAlertShown() bool {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()
	return cyberSecRules.updateAlertShown
}

// GetRaidUnlockString returns the RAID unlock string for Rule 23
func GetRaidUnlockString() string {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()
	return cyberSecRules.raidUnlockString
}

// SetAdWatched marks the ad as watched for Rule 23
func SetAdWatched(watched bool) {
	cyberSecRules.mutex.Lock()
	defer cyberSecRules.mutex.Unlock()
	cyberSecRules.adWatched = watched
}

// IsAdWatched returns whether the ad has been watched
func IsAdWatched() bool {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()
	return cyberSecRules.adWatched
}

// GetBlackSquareCount returns the current count of black squares
func GetBlackSquareCount() int {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()
	return cyberSecRules.blackSquareCount
}

// GenerateBlackSquares creates exactly 5 black squares for Rule 24
func GenerateBlackSquares() string {
	const count = 5 // Always generate exactly 5 black squares

	cyberSecRules.mutex.Lock()
	cyberSecRules.blackSquareCount = count
	cyberSecRules.mutex.Unlock()

	return strings.Repeat("⬛", count)
}

// GetImposterIndices returns the current imposter indices for Rule 25
func GetImposterIndices() []int {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()

	// Return a copy to prevent external modification
	indices := make([]int, len(cyberSecRules.imposterIndices))
	copy(indices, cyberSecRules.imposterIndices)
	return indices
}

// ResetCyberSecurityRules resets all cybersecurity rule states
func ResetCyberSecurityRules() {
	cyberSecRules.mutex.Lock()
	defer cyberSecRules.mutex.Unlock()

	cyberSecRules.updateAlertShown = false
	cyberSecRules.adWatched = false
	cyberSecRules.blackSquareCount = 0
	cyberSecRules.blackboxRuleValidated = false
	cyberSecRules.imposterIndices = []int{}
	cyberSecRules.imposterOriginalChars = []byte{}
	cyberSecRules.imposterRuleValidated = false
	cyberSecRules.lastPasswordLength = 0
}

// CyberSecurityRuleStatus provides status information for cybersecurity rules
type CyberSecurityRuleStatus struct {
	UpdateAlertShown      bool   `json:"update_alert_shown"`
	UpdateString          string `json:"update_string"`
	AdWatched             bool   `json:"ad_watched"`
	RaidUnlockString      string `json:"raid_unlock_string"`
	BlackSquareCount      int    `json:"black_square_count"`
	BlackboxRuleValidated bool   `json:"blackbox_rule_validated"`
	ImposterIndices       []int  `json:"imposter_indices"`
	ImposterOriginalChars []byte `json:"imposter_original_chars"`
	ImposterRuleValidated bool   `json:"imposter_rule_validated"`
}

// GetCyberSecurityStatus returns the current status of all cybersecurity rules
func GetCyberSecurityStatus() CyberSecurityRuleStatus {
	cyberSecRules.mutex.RLock()
	defer cyberSecRules.mutex.RUnlock()

	// Create a copy of the imposterOriginalChars slice
	originalChars := make([]byte, len(cyberSecRules.imposterOriginalChars))
	copy(originalChars, cyberSecRules.imposterOriginalChars)

	return CyberSecurityRuleStatus{
		UpdateAlertShown:      cyberSecRules.updateAlertShown,
		UpdateString:          cyberSecRules.updateString,
		AdWatched:             cyberSecRules.adWatched,
		RaidUnlockString:      cyberSecRules.raidUnlockString,
		BlackSquareCount:      cyberSecRules.blackSquareCount,
		BlackboxRuleValidated: cyberSecRules.blackboxRuleValidated,
		ImposterIndices:       append([]int{}, cyberSecRules.imposterIndices...), // Copy slice
		ImposterOriginalChars: originalChars,
		ImposterRuleValidated: cyberSecRules.imposterRuleValidated,
	}
}
