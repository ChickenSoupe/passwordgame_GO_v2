package fun

import (
	"regexp"
	"strings"
	"time"
	"unicode"
)

// GetRules returns the fun difficulty rules
func GetRules() []Rule {
	return []Rule{
		{
			ID:          1,
			Description: "Your password must include this captcha:",
			Validator:   ValidateCaptcha,
			Hint:        "Enter the 5-digit code shown in the captcha image.",
			HasCaptcha:  true,
		},
		{
			ID:          2,
			Description: "Your password must be at least 10 characters long.",
			Validator:   func(t string) bool { return len(t) >= 10 },
			Hint:        "Add more characters to reach at least 10.",
		},
		{
			ID:          3,
			Description: "Your password must include an uppercase and a lowercase letter.",
			Validator: func(t string) bool {
				hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(t)
				hasLower := regexp.MustCompile(`[a-z]`).MatchString(t)
				return hasUpper && hasLower
			},
			Hint: "Include both UPPERCASE and lowercase letters.",
		},
		{
			ID:          4,
			Description: "Your password must include a special character (!@#$%^&*).",
			Validator: func(t string) bool {
				return regexp.MustCompile(`[!@#$%^&*]`).MatchString(t)
			},
			Hint: "Add one of these: !@#$%^&*",
		},
		{
			ID:          5,
			Description: "Your password must include 'mitochondria' (the powerhouse of the cell). 🦠",
			Validator: func(t string) bool {
				return regexp.MustCompile(`(?i)mitochondria`).MatchString(t)
			},
			Hint: "Include the word 'mitochondria' anywhere in your password.",
		},
		{
			ID:          6,
			Description: "Your password must include the name of a continent.",
			Validator: func(t string) bool {
				continents := []string{"asia", "europe", "africa", "australia", "oceania", "northamerica", "southamerica", "antarctica"}
				lowerT := strings.ToLower(strings.ReplaceAll(t, " ", ""))
				for _, continent := range continents {
					if strings.Contains(lowerT, continent) {
						return true
					}
				}
				return false
			},
			Hint: "Include: Asia, Europe, Africa, Australia, Oceania, North America, South America, or Antarctica.",
		},
		{
			ID:          7,
			Description: "Your password must include a chess piece name.",
			Validator: func(t string) bool {
				pieces := []string{"king", "queen", "rook", "bishop", "knight", "pawn"}
				lowerT := strings.ToLower(t)
				for _, piece := range pieces {
					if strings.Contains(lowerT, piece) {
						return true
					}
				}
				return false
			},
			Hint: "Include: king, queen, rook, bishop, knight, or pawn.",
		},
		{
			ID:          8,
			Description: "Your password must contain the answer to: What is 7 × 8?",
			Validator: func(t string) bool {
				return strings.Contains(t, "56")
			},
			Hint: "Calculate 7 × 8 and include that number.",
		},
		{
			ID:          9,
			Description: "Your password must include an emoji. 🎉",
			Validator: func(t string) bool {
				for _, r := range t {
					if unicode.In(r, unicode.So, unicode.Sm) {
						return true
					}
				}
				return false
			},
			Hint: "Add any emoji to your password! 😊🔥⭐",
		},
		{
			ID:          10,
			Description: "Your password must include a superhero name (superman, batman, spiderman, ironman).",
			Validator: func(t string) bool {
				heroes := []string{"superman", "batman", "spiderman", "ironman", "hulk", "thor", "flash", "wonder"}
				lowerT := strings.ToLower(t)
				for _, hero := range heroes {
					if strings.Contains(lowerT, hero) {
						return true
					}
				}
				return false
			},
			Hint: "Include: superman, batman, spiderman, ironman, hulk, thor, flash, or wonder.",
		},
		{
			ID:          11,
			Description: "Your password must include a programming language name.",
			Validator: func(t string) bool {
				languages := []string{"go", "python", "javascript", "java", "rust", "c++", "php", "ruby", "swift", "kotlin"}
				lowerT := strings.ToLower(t)
				for _, lang := range languages {
					if strings.Contains(lowerT, lang) {
						return true
					}
				}
				return false
			},
			Hint: "Include: go, python, javascript, java, rust, c++, php, ruby, swift, or kotlin.",
		},
		{
			ID:          12,
			Description: "Your password must include a food item (pizza, burger, sushi, taco).",
			Validator: func(t string) bool {
				foods := []string{"pizza", "burger", "sushi", "taco", "pasta", "sandwich", "salad", "soup", "cake", "cookie"}
				lowerT := strings.ToLower(t)
				for _, food := range foods {
					if strings.Contains(lowerT, food) {
						return true
					}
				}
				return false
			},
			Hint: "Include: pizza, burger, sushi, taco, pasta, sandwich, salad, soup, cake, or cookie.",
		},
		{
			ID:          13,
			Description: "Your password must include today's month as a word.",
			Validator: func(t string) bool {
				month := strings.ToLower(time.Now().Format("January"))
				return strings.Contains(strings.ToLower(t), month)
			},
			Hint: "Include the current month: " + time.Now().Format("January"),
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
	HasCaptcha     bool
}
