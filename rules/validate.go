package rules

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"sync"
)

// RuleSet contains a collection of rules for password validation
type RuleSet struct {
	Rules      []Rule
	Difficulty string
}

// Cache for assignments to avoid repeated file reads
var (
	assignmentsCache map[string][]int
	assignmentsMutex sync.RWMutex
	assignmentsLoaded bool
)

// loadAssignments loads assignments.json once and caches it
func loadAssignments() map[string][]int {
	assignmentsMutex.Lock()
	defer assignmentsMutex.Unlock()
	
	if assignmentsLoaded {
		return assignmentsCache
	}

	assignmentsFile, err := os.Open("rules/assignments.json")
	if err != nil {
		log.Printf("Warning: Could not open assignments.json: %v", err)
		assignmentsCache = make(map[string][]int)
		assignmentsLoaded = true
		return assignmentsCache
	}
	defer assignmentsFile.Close()

	var assignments map[string][]int
	if err := json.NewDecoder(assignmentsFile).Decode(&assignments); err != nil {
		log.Printf("Warning: Could not decode assignments.json: %v", err)
		assignmentsCache = make(map[string][]int)
		assignmentsLoaded = true
		return assignmentsCache
	}

	assignmentsCache = assignments
	assignmentsLoaded = true
	return assignmentsCache
}

// NewRuleSet creates a new rule set based on the difficulty level using the pool and assignments.json
func NewRuleSet(difficulty string) *RuleSet {
	var rules []Rule

	// Load assignments from cache
	assignments := loadAssignments()

	// Get rule IDs for the specified difficulty
	ruleIDs, exists := assignments[difficulty]
	if !exists {
		log.Printf("Warning: Difficulty '%s' not found in assignments, using basic", difficulty)
		// fallback: return basic rules from pool
		basicRules := GetRulesByCategory("basic")
		return &RuleSet{Rules: basicRules, Difficulty: difficulty}
	}

	// Get rules from pool by IDs
	rules = GetRulesByIDs(ruleIDs)

	// Sort rules by ID to ensure consistent ordering
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID < rules[j].ID
	})

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

		// Sequential rule visibility logic - Once visible, always visible
		if rs.Rules[i].ID == 1 || i == 0 {
			// Always show rule 1 OR the first rule in the set (even when password is empty)
			rs.Rules[i].IsVisible = true
		} else if oldVisible {
			// Keep visible if was previously visible (don't hide once shown)
			rs.Rules[i].IsVisible = true
		} else if len(password) > 0 && i > 0 {
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

		// Only validate visible rules to improve performance
		if rs.Rules[i].IsVisible {
			rs.Rules[i].IsSatisfied = rs.Rules[i].Validator(password)
			// Mark as newly satisfied if it wasn't satisfied before but is now
			rs.Rules[i].NewlySatisfied = !oldSatisfied && rs.Rules[i].IsSatisfied
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
