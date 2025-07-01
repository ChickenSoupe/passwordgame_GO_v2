package rules

import (
	"regexp"
	"strings"
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

// Pool returns all available rules with unique IDs
func Pool() []Rule {
	return []Rule{
		// Basic Rules (1-10)
		{
			ID:          1,
			Description: "Your password must be at least 8 characters long.",
			Validator:   func(t string) bool { return len(t) >= 8 },
			Hint:        "Add more characters to reach at least 8.",
			Category:    "basic",
		},
		{
			ID:          2,
			Description: "Your password must include an uppercase and a lowercase letter.",
			Validator: func(t string) bool {
				hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(t)
				hasLower := regexp.MustCompile(`[a-z]`).MatchString(t)
				return hasUpper && hasLower
			},
			Hint:     "Include both UPPERCASE and lowercase letters.",
			Category: "basic",
		},
		{
			ID:          3,
			Description: "Your password must include a special character (!@#$%^&*).",
			Validator: func(t string) bool {
				return regexp.MustCompile(`[!@#$%^&*]`).MatchString(t)
			},
			Hint:     "Add one of these: !@#$%^&*",
			Category: "basic",
		},
		{
			ID:          4,
			Description: "Your password must include a number.",
			Validator: func(t string) bool {
				return regexp.MustCompile(`\d`).MatchString(t)
			},
			Hint:     "Add at least one digit (0-9).",
			Category: "basic",
		},
		{
			ID:          5,
			Description: "Your password must not contain common words (password, admin, user).",
			Validator: func(t string) bool {
				lowerT := strings.ToLower(t)
				commonWords := []string{"password", "admin", "user", "login", "welcome"}
				for _, word := range commonWords {
					if strings.Contains(lowerT, word) {
						return false
					}
				}
				return true
			},
			Hint:     "Avoid common words like: password, admin, user, login, welcome.",
			Category: "basic",
		},

		// Intermediate Rules (11-20)
		{
			ID:          11,
			Description: "Your password must be at least 12 characters long.",
			Validator:   func(t string) bool { return len(t) >= 12 },
			Hint:        "Add more characters to reach at least 12.",
			Category:    "intermediate",
		},
		{
			ID:          12,
			Description: "Your password must include at least 2 numbers.",
			Validator: func(t string) bool {
				count := 0
				for _, char := range t {
					if unicode.IsDigit(char) {
						count++
					}
				}
				return count >= 2
			},
			Hint:     "Add at least 2 digits (0-9).",
			Category: "intermediate",
		},
		{
			ID:          13,
			Description: "Your password must include at least 2 special characters.",
			Validator: func(t string) bool {
				count := 0
				specialChars := "!@#$%^&*"
				for _, char := range t {
					if strings.ContainsRune(specialChars, char) {
						count++
					}
				}
				return count >= 2
			},
			Hint:     "Add at least 2 special characters: !@#$%^&*",
			Category: "intermediate",
		},
		{
			ID:          14,
			Description: "Your password must not contain repeated characters (e.g., 'aa', '11').",
			Validator: func(t string) bool {
				for i := 0; i < len(t)-1; i++ {
					if t[i] == t[i+1] {
						return false
					}
				}
				return true
			},
			Hint:     "Remove any repeated consecutive characters.",
			Category: "intermediate",
		},
		{
			ID:          15,
			Description: "Your password must contain at least one vowel (a, e, i, o, u).",
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

		// Hard Rules (21-30)
		{
			ID:          21,
			Description: "Your password must be at least 16 characters long.",
			Validator:   func(t string) bool { return len(t) >= 16 },
			Hint:        "Add more characters to reach at least 16.",
			Category:    "hard",
		},
		{
			ID:          22,
			Description: "Your password must include at least 3 uppercase letters.",
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
			Category: "hard",
		},
		{
			ID:          23,
			Description: "Your password must include at least 3 numbers.",
			Validator: func(t string) bool {
				count := 0
				for _, char := range t {
					if unicode.IsDigit(char) {
						count++
					}
				}
				return count >= 3
			},
			Hint:     "Add at least 3 digits (0-9).",
			Category: "hard",
		},
		{
			ID:          24,
			Description: "Your password must not contain any dictionary words.",
			Validator: func(t string) bool {
				lowerT := strings.ToLower(t)
				dictWords := []string{"password", "admin", "user", "login", "welcome", "hello", "world", "computer", "internet", "security", "system", "access", "account", "email", "phone", "address", "name", "birthday", "date", "time", "year", "month", "day"}
				for _, word := range dictWords {
					if strings.Contains(lowerT, word) {
						return false
					}
				}
				return true
			},
			Hint:     "Avoid common dictionary words.",
			Category: "hard",
		},
		{
			ID:          25,
			Description: "Your password must contain at least one character from each: uppercase, lowercase, number, special character.",
			Validator: func(t string) bool {
				hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(t)
				hasLower := regexp.MustCompile(`[a-z]`).MatchString(t)
				hasDigit := regexp.MustCompile(`\d`).MatchString(t)
				hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(t)
				return hasUpper && hasLower && hasDigit && hasSpecial
			},
			Hint:     "Include at least one: UPPERCASE, lowercase, digit, special character.",
			Category: "hard",
		},

		// Fun/Creative Rules (31-40)
		{
			ID:          31,
			Description: "Your password must include this captcha:",
			Validator:   ValidateCaptcha,
			Hint:        "Enter the 5-digit code shown in the captcha image.",
			HasCaptcha:  true,
			Category:    "fun",
		},
		{
			ID:          32,
			Description: "Your password must include today's Wordle answer. 🎯",
			Validator:   ValidateWordleAnswer,
			Hint:        "Include today's Wordle solution: " + GetTodaysAnswerForHint(),
			Category:    "fun",
		},
		{
			ID:          33,
			Description: "Your password must include the name of a month.",
			Validator: func(t string) bool {
				lowerT := strings.ToLower(t)
				months := []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}
				for _, month := range months {
					if strings.Contains(lowerT, month) {
						return true
					}
				}
				return false
			},
			Hint:     "Include a month name (january, february, etc.).",
			Category: "fun",
		},
		{
			ID:          34,
			Description: "Your password must include a color name.",
			Validator: func(t string) bool {
				lowerT := strings.ToLower(t)
				colors := []string{"red", "blue", "green", "yellow", "orange", "purple", "pink", "black", "white", "brown", "gray", "grey"}
				for _, color := range colors {
					if strings.Contains(lowerT, color) {
						return true
					}
				}
				return false
			},
			Hint:     "Include a color name (red, blue, green, etc.).",
			Category: "fun",
		},
		{
			ID:          35,
			Description: "Your password must include an animal name.",
			Validator: func(t string) bool {
				lowerT := strings.ToLower(t)
				animals := []string{"cat", "dog", "bird", "fish", "lion", "tiger", "bear", "wolf", "fox", "rabbit", "mouse", "elephant", "giraffe", "zebra", "horse", "cow", "pig", "sheep", "goat", "chicken"}
				for _, animal := range animals {
					if strings.Contains(lowerT, animal) {
						return true
					}
				}
				return false
			},
			Hint:     "Include an animal name (cat, dog, lion, etc.).",
			Category: "fun",
		},
		{
			ID:          36,
			Description: "Your password must include the digits of the current year (2024).",
			Validator: func(t string) bool {
				return strings.Contains(t, "2024")
			},
			Hint:     "Include the current year: 2024",
			Category: "fun",
		},
		{
			ID:          37,
			Description: "Your password must include a mathematical operation (+, -, *, /).",
			Validator: func(t string) bool {
				mathOps := "+-*/"
				for _, char := range t {
					if strings.ContainsRune(mathOps, char) {
						return true
					}
				}
				return false
			},
			Hint:     "Include a math operator: +, -, *, /",
			Category: "fun",
		},

		// Expert Rules (41-50)
		{
			ID:          41,
			Description: "Your password must be at least 20 characters long.",
			Validator:   func(t string) bool { return len(t) >= 20 },
			Hint:        "Add more characters to reach at least 20.",
			Category:    "expert",
		},
		{
			ID:          42,
			Description: "Your password must include at least 4 different special characters.",
			Validator: func(t string) bool {
				specialChars := "!@#$%^&*"
				found := make(map[rune]bool)
				for _, char := range t {
					if strings.ContainsRune(specialChars, char) {
						found[char] = true
					}
				}
				return len(found) >= 4
			},
			Hint:     "Include at least 4 different special characters: !@#$%^&*",
			Category: "expert",
		},
		{
			ID:          43,
			Description: "Your password must contain a palindrome (word that reads the same forwards and backwards).",
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
		{
			ID:          44,
			Description: "Your password must include Roman numerals (I, V, X, L, C, D, M).",
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
			Category: "expert",
		},
		{
			ID:          45,
			Description: "Your password must include a prime number (2, 3, 5, 7, 11, 13, etc.).",
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
			Category: "expert",
		},
	}
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