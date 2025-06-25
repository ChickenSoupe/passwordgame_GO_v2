package rules

import (
	"passgame/rules/basic"
	"passgame/rules/fun"
	"passgame/rules/hard"
	"passgame/rules/intermediate"
)

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

// RuleSet contains a collection of rules for password validation
type RuleSet struct {
	Rules      []Rule
	Difficulty string
}

// NewRuleSet creates a new rule set based on the difficulty level
func NewRuleSet(difficulty string) *RuleSet {
	var rules []Rule

	switch difficulty {
	case "basic":
		basicRules := basic.GetRules()
		for _, r := range basicRules {
			rules = append(rules, Rule{
				ID:          r.ID,
				Description: r.Description,
				Validator:   r.Validator,
				Hint:        r.Hint,
			})
		}
	case "intermediate":
		intermediateRules := intermediate.GetRules()
		for _, r := range intermediateRules {
			rules = append(rules, Rule{
				ID:          r.ID,
				Description: r.Description,
				Validator:   r.Validator,
				Hint:        r.Hint,
			})
		}
	case "hard":
		hardRules := hard.GetRules()
		for _, r := range hardRules {
			rules = append(rules, Rule{
				ID:          r.ID,
				Description: r.Description,
				Validator:   r.Validator,
				Hint:        r.Hint,
			})
		}
	case "fun":
		funRules := fun.GetRules()
		for _, r := range funRules {
			rules = append(rules, Rule{
				ID:          r.ID,
				Description: r.Description,
				Validator:   r.Validator,
				Hint:        r.Hint,
				HasCaptcha:  r.HasCaptcha,
			})
		}
	default:
		// Default to basic if unknown difficulty
		basicRules := basic.GetRules()
		for _, r := range basicRules {
			rules = append(rules, Rule{
				ID:          r.ID,
				Description: r.Description,
				Validator:   r.Validator,
				Hint:        r.Hint,
			})
		}
	}

	return &RuleSet{
		Rules:      rules,
		Difficulty: difficulty,
	}
}

// ValidatePassword validates the password against all rules in the rule set
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

		// Sequential rule visibility logic - Once visible, always visible
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
				if visibleRules[i].ID > visibleRules[j].ID {
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
