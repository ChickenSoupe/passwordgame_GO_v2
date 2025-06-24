package basic

import (
	"regexp"
	"strings"
)

// GetRules returns the basic difficulty rules
func GetRules() []Rule {
	return []Rule{
		{
			ID:          1,
			Description: "Your password must be at least 8 characters long.",
			Validator:   func(t string) bool { return len(t) >= 8 },
			Hint:        "Add more characters to reach at least 8.",
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
			Description: "Your password must include a number.",
			Validator: func(t string) bool {
				return regexp.MustCompile(`\d`).MatchString(t)
			},
			Hint: "Add at least one digit (0-9).",
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
			Hint: "Avoid common words like: password, admin, user, login, welcome.",
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
