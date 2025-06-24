package intermediate

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GetRules returns the intermediate difficulty rules
func GetRules() []Rule {
	return []Rule{
		{
			ID:          1,
			Description: "Your password must be at least 12 characters long.",
			Validator:   func(t string) bool { return len(t) >= 12 },
			Hint:        "Add more characters to reach at least 12.",
		},
		{
			ID:          2,
			Description: "Your password must include an uppercase and a lowercase letter.",
			Validator: func(t string) bool {
				hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(t)
				hasLower := regexp.MustCompile(`[a-z]`).MatchString(t)
				return hasUpper && hasLower
			},
			Hint: "Include both UPPERCASE and lowercase letters.",
		},
		{
			ID:          3,
			Description: "Your password must include a special character (!@#$%^&*).",
			Validator: func(t string) bool {
				return regexp.MustCompile(`[!@#$%^&*]`).MatchString(t)
			},
			Hint: "Add one of these: !@#$%^&*",
		},
		{
			ID:          4,
			Description: "Your password must include a 2-digit number.",
			Validator: func(t string) bool {
				return regexp.MustCompile(`\d{2}`).MatchString(t)
			},
			Hint: "Include at least two consecutive digits (e.g., 23, 45).",
		},
		{
			ID:          5,
			Description: "Your password must contain all English vowels (a, e, i, o, u).",
			Validator: func(t string) bool {
				vowels := []string{"a", "e", "i", "o", "u"}
				for _, vowel := range vowels {
					if !regexp.MustCompile(`(?i)` + vowel).MatchString(t) {
						return false
					}
				}
				return true
			},
			Hint: "Make sure to include: a, e, i, o, u (case doesn't matter).",
		},
		{
			ID:          6,
			Description: "Your password must include a 2-digit prime number.",
			Validator: func(t string) bool {
				primes := []string{"11", "13", "17", "19", "23", "29", "31", "37", "41", "43", "47", "53", "59", "61", "67", "71", "73", "79", "83", "89", "97"}
				for _, prime := range primes {
					if strings.Contains(t, prime) {
						return true
					}
				}
				return false
			},
			Hint: "Include a 2-digit prime like: 11, 13, 17, 19, 23, 29, etc.",
		},
		{
			ID:          7,
			Description: "The digits in your password must sum to 25.",
			Validator: func(t string) bool {
				sum := 0
				for _, char := range t {
					if digit, err := strconv.Atoi(string(char)); err == nil && digit >= 0 {
						sum += digit
					}
				}
				return sum == 25
			},
			Hint: "Make sure all digits in your password add up to exactly 25.",
		},
		{
			ID:          8,
			Description: "Your password must include today's month as a word.",
			Validator: func(t string) bool {
				month := strings.ToLower(time.Now().Format("January"))
				return strings.Contains(strings.ToLower(t), month)
			},
			Hint: "Include the current month: " + time.Now().Format("January"),
		},
		{
			ID:          9,
			Description: "Your password must contain a Roman numeral (I, V, X, L, C, D, M).",
			Validator: func(t string) bool {
				return regexp.MustCompile(`[IVXLCDM]`).MatchString(strings.ToUpper(t))
			},
			Hint: "Include a Roman numeral: I, V, X, L, C, D, or M.",
		},
		{
			ID:          10,
			Description: "Your password must include its own length as a number.",
			Validator: func(t string) bool {
				length := strconv.Itoa(len(t))
				return strings.Contains(t, length)
			},
			Hint: "If your password is 25 characters long, it must contain '25'.",
		},
	}
}

// Rule represents a password validation rule
type Rule struct {
	ID             int
	Description    string
	Validator      func(string) bool
	IsSatisfied    bool
	Hint           string
	NewlyRevealed  bool
	NewlySatisfied bool
	IsVisible      bool
}
