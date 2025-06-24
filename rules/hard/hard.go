package hard

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GetRules returns the hard difficulty rules
func GetRules() []Rule {
	return []Rule{
		{
			ID:          1,
			Description: "Your password must be at least 16 characters long.",
			Validator:   func(t string) bool { return len(t) >= 16 },
			Hint:        "Add more characters to reach at least 16.",
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
			Description: "Your password must include at least 2 different special characters (!@#$%^&*).",
			Validator: func(t string) bool {
				specialChars := "!@#$%^&*"
				foundChars := make(map[rune]bool)
				for _, char := range t {
					if strings.ContainsRune(specialChars, char) {
						foundChars[char] = true
					}
				}
				return len(foundChars) >= 2
			},
			Hint: "Use at least 2 different special characters from: !@#$%^&*",
		},
		{
			ID:          4,
			Description: "Your password must include a negative number.",
			Validator: func(t string) bool {
				return regexp.MustCompile(`-\d`).MatchString(t)
			},
			Hint: "Include a minus sign followed by a digit (e.g., -5).",
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
			Description: "The digits in your password must sum to 30.",
			Validator: func(t string) bool {
				sum := 0
				for _, char := range t {
					if digit, err := strconv.Atoi(string(char)); err == nil && digit >= 0 {
						sum += digit
					}
				}
				return sum == 30
			},
			Hint: "Make sure all digits in your password add up to exactly 30.",
		},
		{
			ID:          8,
			Description: "Your password must contain π (pi) to 5 decimal places: 3.14159",
			Validator: func(t string) bool {
				return strings.Contains(t, "3.14159")
			},
			Hint: "Include exactly: 3.14159",
		},
		{
			ID:          9,
			Description: "Your password must have equal numbers of vowels and consonants.",
			Validator: func(t string) bool {
				vowelCount := len(regexp.MustCompile(`[aeiouAEIOU]`).FindAllString(t, -1))
				consonantCount := len(regexp.MustCompile(`[bcdfghjklmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ]`).FindAllString(t, -1))
				return vowelCount == consonantCount && vowelCount > 0
			},
			Hint: "Balance the vowels (a,e,i,o,u) and consonants equally.",
		},
		{
			ID:          10,
			Description: "Your password must include today's month as a word.",
			Validator: func(t string) bool {
				month := strings.ToLower(time.Now().Format("January"))
				return strings.Contains(strings.ToLower(t), month)
			},
			Hint: "Include the current month: " + time.Now().Format("January"),
		},
		{
			ID:          11,
			Description: "Your password must contain a Roman numeral (I, V, X, L, C, D, M).",
			Validator: func(t string) bool {
				return regexp.MustCompile(`[IVXLCDM]`).MatchString(strings.ToUpper(t))
			},
			Hint: "Include a Roman numeral: I, V, X, L, C, D, or M.",
		},
		{
			ID:          12,
			Description: "Your password must include a color in hex format (#RRGGBB).",
			Validator: func(t string) bool {
				return regexp.MustCompile(`#[0-9A-Fa-f]{6}`).MatchString(t)
			},
			Hint: "Include a hex color like #FF0000 (red) or #00FF00 (green).",
		},
		{
			ID:          13,
			Description: "Your password must end with the current year.",
			Validator: func(t string) bool {
				year := strconv.Itoa(time.Now().Year())
				return strings.HasSuffix(t, year)
			},
			Hint: "End your password with: " + strconv.Itoa(time.Now().Year()),
		},
		{
			ID:          14,
			Description: "Your password must include its own length as a number.",
			Validator: func(t string) bool {
				length := strconv.Itoa(len(t))
				return strings.Contains(t, length)
			},
			Hint: "If your password is 25 characters long, it must contain '25'.",
		},
		{
			ID:          15,
			Description: "Your password must include the chemical symbol for gold (Au).",
			Validator: func(t string) bool {
				return strings.Contains(t, "Au")
			},
			Hint: "Include 'Au' - the chemical symbol for gold.",
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
