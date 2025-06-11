package basic

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Rule struct {
	ID             int
	Description    string
	Validator      func(string) bool
	IsSatisfied    bool
	Hint           string
	NewlyRevealed  bool
	NewlySatisfied bool // Track when rule becomes satisfied for the first time
	IsVisible      bool // Track if rule has ever been shown
}

type RuleSet struct {
	Rules []Rule
}

func NewRuleSet() *RuleSet {
	return &RuleSet{
		Rules: []Rule{
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
				Description: "The digits in your password must sum to 25.",
				Validator: func(t string) bool {
					sum := 0
					for _, char := range t {
						if digit, err := strconv.Atoi(string(char)); err == nil {
							sum += digit
						}
					}
					return sum == 25
				},
				Hint: "Make sure all digits in your password add up to exactly 25.",
			},
			{
				ID:          8,
				Description: "Your password must include 'mitochondria' (the powerhouse of the cell). 🦠",
				Validator: func(t string) bool {
					return regexp.MustCompile(`(?i)mitochondria`).MatchString(t)
				},
				Hint: "Include the word 'mitochondria' anywhere in your password.",
			},
			{
				ID:          9,
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
				ID:          10,
				Description: "Your password must contain π (pi) to 5 decimal places: 3.14159",
				Validator: func(t string) bool {
					return strings.Contains(t, "3.14159")
				},
				Hint: "Include exactly: 3.14159",
			},
			{
				ID:          11,
				Description: "Your password must have equal numbers of vowels and consonants.",
				Validator: func(t string) bool {
					vowelCount := len(regexp.MustCompile(`[aeiouAEIOU]`).FindAllString(t, -1))
					consonantCount := len(regexp.MustCompile(`[bcdfghjklmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ]`).FindAllString(t, -1))
					return vowelCount == consonantCount && vowelCount > 0
				},
				Hint: "Balance the vowels (a,e,i,o,u) and consonants equally.",
			},
			{
				ID:          12,
				Description: "Your password must include its own length as a number.",
				Validator: func(t string) bool {
					length := strconv.Itoa(len(t))
					return strings.Contains(t, length)
				},
				Hint: "If your password is 25 characters long, it must contain '25'.",
			},
			{
				ID:          13,
				Description: "Your password must include today's month as a word.",
				Validator: func(t string) bool {
					month := strings.ToLower(time.Now().Format("January"))
					return strings.Contains(strings.ToLower(t), month)
				},
				Hint: fmt.Sprintf("Include the current month: %s", time.Now().Format("January")),
			},
			{
				ID:          14,
				Description: "Your password must contain a Roman numeral (I, V, X, L, C, D, M).",
				Validator: func(t string) bool {
					return regexp.MustCompile(`[IVXLCDM]`).MatchString(strings.ToUpper(t))
				},
				Hint: "Include a Roman numeral: I, V, X, L, C, D, or M.",
			},
			{
				ID:          15,
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
				ID:          16,
				Description: "Your password must contain the chemical symbol for gold (Au).",
				Validator: func(t string) bool {
					return strings.Contains(t, "Au")
				},
				Hint: "Include 'Au' - the chemical symbol for gold.",
			},
			{
				ID:          17,
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
				ID:          18,
				Description: "Your password must contain the answer to: What is 7 × 8?",
				Validator: func(t string) bool {
					return strings.Contains(t, "56")
				},
				Hint: "Calculate 7 × 8 and include that number.",
			},
			{
				ID:          19,
				Description: "Your password must include a color in hex format (#RRGGBB).",
				Validator: func(t string) bool {
					return regexp.MustCompile(`#[0-9A-Fa-f]{6}`).MatchString(t)
				},
				Hint: "Include a hex color like #FF0000 (red) or #00FF00 (green).",
			},
			{
				ID:          20,
				Description: "Your password must end with the current year.",
				Validator: func(t string) bool {
					year := strconv.Itoa(time.Now().Year())
					return strings.HasSuffix(t, year)
				},
				Hint: fmt.Sprintf("End your password with: %d", time.Now().Year()),
			},
		},
	}
}

// ValidatePassword validates the password against all rules
func ValidatePassword(rs *RuleSet, password string, previousStates []bool, previousVisible []bool) {
	for i := range rs.Rules {
		oldSatisfied := false
		oldVisible := false
		if i < len(previousStates) {
			oldSatisfied = previousStates[i]
		}
		if i < len(previousVisible) {
			oldVisible = previousVisible[i]
		}

		rs.Rules[i].IsSatisfied = rs.Rules[i].Validator(password)

		// Mark as newly satisfied if it wasn't satisfied before but is now
		rs.Rules[i].NewlySatisfied = !oldSatisfied && rs.Rules[i].IsSatisfied

		// Sequential rule visibility logic - MODIFIED: Once visible, always visible
		if rs.Rules[i].ID == 1 {
			// Always show rule 1 (even when password is empty)
			rs.Rules[i].IsVisible = true
		} else if oldVisible {
			// Keep visible if was previously visible (don't hide once shown)
			rs.Rules[i].IsVisible = true
		} else if len(password) > 0 {
			// For rules 2 and above, check if all previous rules are visible
			allPreviousVisible := true
			for j := 0; j < i; j++ {
				if !rs.Rules[j].IsVisible {
					allPreviousVisible = false
					break
				}
			}

			// Show this rule only if all previous rules are visible AND the immediately previous rule is satisfied
			if allPreviousVisible && rs.Rules[i-1].IsSatisfied {
				rs.Rules[i].IsVisible = true
			}
		}

		// Mark as newly revealed if it wasn't visible before but is now
		rs.Rules[i].NewlyRevealed = !oldVisible && rs.Rules[i].IsVisible
	}
}

// GetSatisfiedCount returns the number of satisfied rules
func GetSatisfiedCount(rs *RuleSet) int {
	count := 0
	for _, rule := range rs.Rules {
		if rule.IsSatisfied {
			count++
		}
	}
	return count
}

// GetSatisfiedStates returns a slice of boolean values indicating which rules are satisfied
func GetSatisfiedStates(rs *RuleSet) []bool {
	states := make([]bool, len(rs.Rules))
	for i, rule := range rs.Rules {
		states[i] = rule.IsSatisfied
	}
	return states
}

// GetVisibleStates returns a slice of boolean values indicating which rules are visible
func GetVisibleStates(rs *RuleSet) []bool {
	states := make([]bool, len(rs.Rules))
	for i, rule := range rs.Rules {
		states[i] = rule.IsVisible
	}
	return states
}

// GetSortedVisibleRules returns visible rules sorted with unsatisfied rules first, then satisfied rules
func GetSortedVisibleRules(rs *RuleSet) []Rule {
	var visibleRules []Rule

	// Collect only visible rules
	for _, rule := range rs.Rules {
		if rule.IsVisible {
			visibleRules = append(visibleRules, rule)
		}
	}

	// Sort: unsatisfied rules first (by ID ascending), then satisfied rules (by ID ascending)
	for i := 0; i < len(visibleRules)-1; i++ {
		for j := i + 1; j < len(visibleRules); j++ {
			// If both have same satisfaction status, sort by ID ascending (lowest to highest)
			if visibleRules[i].IsSatisfied == visibleRules[j].IsSatisfied {
				if visibleRules[i].ID < visibleRules[j].ID {
					visibleRules[i], visibleRules[j] = visibleRules[j], visibleRules[i]
				}
			} else {
				// Unsatisfied rules come first
				if visibleRules[i].IsSatisfied && !visibleRules[j].IsSatisfied {
					visibleRules[i], visibleRules[j] = visibleRules[j], visibleRules[i]
				}
			}
		}
	}

	return visibleRules
}
