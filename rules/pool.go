package rules

import (
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

// Rule represents a password validation rule
type Rule struct {
	ID             int               `json:"id"`
	Description    string            `json:"description"`
	Validator      func(string) bool `json:"-"`
	IsSatisfied    bool              `json:"is_satisfied"`
	Hint           string            `json:"hint"`
	NewlyRevealed  bool              `json:"newly_revealed"`
	NewlySatisfied bool              `json:"newly_satisfied"`
	IsVisible      bool              `json:"is_visible"`
	HasCaptcha     bool              `json:"has_captcha"`
	Category       string            `json:"category"`
}

// Cache for the rule pool
var (
	rulePool   []Rule
	poolMutex  sync.RWMutex
	poolLoaded bool
)

// Pool returns all available rules with unique IDs
func Pool() []Rule {
	poolMutex.Lock()
	defer poolMutex.Unlock()

	if poolLoaded {
		return rulePool
	}

	rulePool = []Rule{
		// Rule 1: Must be at least 8 characters long
		{
			ID:          1,
			Description: "Must be at least 8 characters long",
			Validator:   func(t string) bool { return len(t) >= 8 },
			Hint:        "Add more characters to reach at least 8.",
			Category:    "basic",
		},
		// Rule 2: Must include both uppercase and lowercase letters
		{
			ID:          2,
			Description: "Must include both uppercase and lowercase letters",
			Validator: func(t string) bool {
				hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(t)
				hasLower := regexp.MustCompile(`[a-z]`).MatchString(t)
				return hasUpper && hasLower
			},
			Hint:     "Include both UPPERCASE and lowercase letters.",
			Category: "basic",
		},
		// Rule 3: Must include a special character (!@#$%^&*)
		{
			ID:          3,
			Description: "Must include a special character (!@#$%^&*)",
			Validator: func(t string) bool {
				return regexp.MustCompile(`[!@#$%^&*\\]`).MatchString(t)
			},
			Hint:     "Add one of these: !@#$%^&*\\",
			Category: "basic",
		},
		// Rule 4: Must include a number
		{
			ID:          4,
			Description: "Must include a number",
			Validator: func(t string) bool {
				return regexp.MustCompile(`\d`).MatchString(t)
			},
			Hint:     "Add at least one digit (0-9).",
			Category: "basic",
		},
		// Rule 5: Must include Roman numerals (I, V, X, L, C, D, M)
		{
			ID:          5,
			Description: "Must include Roman numerals (I, V, X, L, C, D, M)",
			Validator: func(t string) bool {
				romanNumerals := "IVXLCDM"
				for _, char := range t {
					if strings.ContainsRune(romanNumerals, char) {
						return true
					}
				}
				return false
			},
			Hint:     "Include Roman numerals: I, V, X, L, C, D, M",
			Category: "basic",
		},
		// Rule 6: Must include a prime number
		{
			ID:          6,
			Description: "Must include a prime number",
			Validator: func(t string) bool {
				primes := []string{"2", "3", "5", "7", "11", "13", "17", "19", "23", "29", "31", "37", "41", "43", "47"}
				for _, prime := range primes {
					if strings.Contains(t, prime) {
						return true
					}
				}
				return false
			},
			Hint:     "Include a prime number: 2, 3, 5, 7, 11, 13, etc.",
			Category: "basic",
		},
		// Rule 7: Must contain the current day of the week
		{
			ID:          7,
			Description: "Must contain the current day of the week",
			Validator: func(t string) bool {
				currentDay := strings.ToLower(time.Now().Weekday().String())
				return strings.Contains(strings.ToLower(t), currentDay)
			},
			Hint:     "Include today's day of the week: " + time.Now().Weekday().String(),
			Category: "intermediate",
		},
		// Rule 8: Must contain one of our following sponsors: (Pepsi, Starbucks, Shell)
		{
			ID:          8,
			Description: "Must contain one of our following sponsors: (Pepsi, Starbucks, Shell)",
			Validator: func(t string) bool {
				lowerT := strings.ToLower(t)
				sponsors := []string{"pepsi", "starbucks", "shell"}
				for _, sponsor := range sponsors {
					if strings.Contains(lowerT, sponsor) {
						return true
					}
				}
				return false
			},
			Hint:     "Include one of our sponsors: Pepsi, Starbucks, Shell",
			Category: "intermediate",
		},
		// Rule 9: Must contain at least one vowel
		{
			ID:          9,
			Description: "Must contain at least one vowel",
			Validator: func(t string) bool {
				vowels := "aeiouAEIOU"
				for _, char := range t {
					if strings.ContainsRune(vowels, char) {
						return true
					}
				}
				return false
			},
			Hint:     "Add at least one vowel: a, e, i, o, u",
			Category: "intermediate",
		},
		// Rule 10: Must include the current month name
		{
			ID:          10,
			Description: "Must include the current month name",
			Validator: func(t string) bool {
				currentMonth := strings.ToLower(time.Now().Month().String())
				return strings.Contains(strings.ToLower(t), currentMonth)
			},
			Hint:     "Include the current month: " + time.Now().Month().String(),
			Category: "intermediate",
		},
		// Rule 11: Must be at least 16 characters long
		{
			ID:          11,
			Description: "Must be at least 16 characters long",
			Validator:   func(t string) bool { return len(t) >= 16 },
			Hint:        "Add more characters to reach at least 16.",
			Category:    "intermediate",
		},
		// Rule 12: Must include at least 3 uppercase letters
		{
			ID:          12,
			Description: "Must include at least 3 uppercase letters",
			Validator: func(t string) bool {
				count := 0
				for _, char := range t {
					if unicode.IsUpper(char) {
						count++
					}
				}
				return count >= 3
			},
			Hint:     "Add at least 3 UPPERCASE letters.",
			Category: "intermediate",
		},
		// Rule 13: Must include the first 3 numbers of a mathematical constant: random
		{
			ID:          13,
			Description: "Must include the first 3 numbers of the following mathematical constant: random",
			Validator:   ValidateMathConstant,
			Hint: func() string {
				return "Include the first 3 digits of " + GetMathConstantForHint()
			}(),
			Category: "hard",
		},
		// Rule 14: Update alert box
		{
			ID:          14,
			Description: "A new password rule just got updated! Please click update on the alertbox!",
			Validator:   Rule14UpdateAlert,
			Hint:        "After the update, include '" + GetUpdateString() + "' in your password.",
			Category:    "expert",
		},
		// Rule 15: Must include a captcha (5-digit code)
		{
			ID:          15,
			Description: "Must include a captcha (5-digit code)",
			Validator:   ValidateCaptcha,
			Hint:        "Enter the 5-digit code shown in the captcha image.",
			HasCaptcha:  true,
			Category:    "hard",
		},
		// Rule 16: Must include today's Wordle answer
		{
			ID:          16,
			Description: "Must include today's Wordle answer",
			Validator:   ValidateWordleAnswer,
			Hint:        "Include today's Wordle solution: " + GetTodaysAnswerForHint(),
			Category:    "hard",
		},
		// Rule 17: Must include the word in this QR code
		{
			ID:          17,
			Description: "Must include the word in this QR code",
			Validator:   ValidateQRCodeWord,
			HasCaptcha:  true,
			Hint:        "Scan the QR code to get the required word.",
			Category:    "hard",
		},
		// Rule 18: Must include a Hex code of the following color
		{
			ID:          18,
			Description: "Must include a Hex code of the following color",
			Validator:   ValidateHexColor,
			Hint: func() string {
				return "Include the hex color code for " + GetColorForHint()
			}(),
			HasCaptcha: true, // We'll use the captcha display logic to show the color
			Category:   "hard",
		},
		// Rule 19: Must include the best chess move
		{
			ID:          19,
			Description: "Must include the best chess move (image)",
			Validator:   ValidateChessMove,
			Hint: func() string {
				_, bestMove := GetCurrentChessPosition()
				if bestMove == "" {
					return "Analyzing chess position..."
				}
				return "Best move: " + bestMove
			}(),
			HasCaptcha: true, // Reuse captcha display logic for chess board
			Category:   "expert",
		},
		// Rule 20: Your password is not strong enough 🏋️
		{
			ID:          20,
			Description: "Your password is not strong enough 🏋️",
			Validator: func(t string) bool {
				// Satisfy if there are at least 3 🏋️ emojis
				emoji := "🏋️"
				count := strings.Count(t, emoji)
				return count >= 3
			},
			Hint:     "Add at least 3 🏋️ emojis to your password.",
			Category: "expert",
		},
		// Rule 21: Must contain a palindrome (3+ characters)
		{
			ID:          21,
			Description: "Must contain a palindrome (3+ characters)",
			Validator: func(t string) bool {
				// Check for palindromes of length 3 or more
				for i := 0; i < len(t); i++ {
					for j := i + 3; j <= len(t); j++ {
						substr := t[i:j]
						if isPalindrome(substr) {
							return true
						}
					}
				}
				return false
			},
			Hint:     "Include a palindrome like 'aba', 'racecar', or '121'.",
			Category: "expert",
		},
		// Rule 22: Must include "pdf file"
		{
			ID:          22,
			Description: "Must include \"pdf file\" (link to malware, when just need the word pdf file)",
			Validator:   Rule22PDFFile,
			Hint:        "Include the phrase 'pdf file' in your password.",
			Category:    "expert",
		},
		// Rule 23: Locks password textbox
		{
			ID:          23,
			Description: "_Locks password textbox_ Oh no! Your password textbox is locked! Watch this raid shadows legend ad to unlock your textbox!",
			Validator:   Rule23PasswordLock,
			Hint:        "After the ad, include '" + GetRaidUnlockString() + "' in your password.",
			Category:    "expert",
		},
		// Rule 24: Ransomware attack warning
		{
			ID:          24,
			Description: "!!Warning!! a ransomware attack is trying to get your password, delete the blackbox to defend it!",
			Validator:   Rule24RansomwareAttack,
			Hint:        "Delete the black squares to defend your password!",
			Category:    "expert",
		},
		// Rule 25: Insider threat detection
		{
			ID:          25,
			Description: "It seems like someone here leaked your information, find the insider threat in your password!",
			Validator:   Rule25InsiderThreat,
			Hint:        "Delete the imposter letters (highlighted in red) from your password! Add 'NOIMPOSTER' to your password when done.",
			Category:    "expert",
		},
	}

	poolLoaded = true
	return rulePool
}

// Helper function to check if a string is a palindrome
func isPalindrome(s string) bool {
	s = strings.ToLower(s)
	for i := 0; i < len(s)/2; i++ {
		if s[i] != s[len(s)-1-i] {
			return false
		}
	}
	return true
}

// GetRuleByID returns a rule by its ID from the pool
func GetRuleByID(id int) *Rule {
	pool := Pool()
	for _, rule := range pool {
		if rule.ID == id {
			return &rule
		}
	}
	return nil
}

// GetRulesByCategory returns all rules in a specific category
func GetRulesByCategory(category string) []Rule {
	pool := Pool()
	var categoryRules []Rule
	for _, rule := range pool {
		if rule.Category == category {
			categoryRules = append(categoryRules, rule)
		}
	}
	return categoryRules
}

// GetRulesByIDs returns rules matching the provided IDs
func GetRulesByIDs(ids []int) []Rule {
	pool := Pool()
	idSet := make(map[int]struct{})
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	var matchingRules []Rule
	for _, rule := range pool {
		if _, exists := idSet[rule.ID]; exists {
			matchingRules = append(matchingRules, rule)
		}
	}
	return matchingRules
}
