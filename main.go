package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"sort"
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
}

type PageData struct {
	Password string
	Rules    []Rule
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

func (rs *RuleSet) ValidatePassword(password string, previousStates []bool) {
	for i := range rs.Rules {
		oldSatisfied := false
		if i < len(previousStates) {
			oldSatisfied = previousStates[i]
		}

		rs.Rules[i].IsSatisfied = rs.Rules[i].Validator(password)

		// Mark as newly satisfied if it wasn't satisfied before but is now
		rs.Rules[i].NewlySatisfied = !oldSatisfied && rs.Rules[i].IsSatisfied
	}
}

func (rs *RuleSet) GetSatisfiedCount() int {
	count := 0
	for _, rule := range rs.Rules {
		if rule.IsSatisfied {
			count++
		}
	}
	return count
}

func (rs *RuleSet) GetSatisfiedStates() []bool {
	states := make([]bool, len(rs.Rules))
	for i, rule := range rs.Rules {
		states[i] = rule.IsSatisfied
	}
	return states
}

// Sort rules: unsatisfied rules first (highest ID to lowest), then satisfied rules (highest ID to lowest)
func (rs *RuleSet) GetSortedVisibleRules(maxVisible int) []Rule {
	var visibleRules []Rule

	// Collect visible rules (either satisfied or within maxVisible range)
	for i, rule := range rs.Rules {
		if rule.IsSatisfied || i < maxVisible {
			visibleRules = append(visibleRules, rule)
		}
	}

	// Sort: unsatisfied rules first (reverse ID order), then satisfied rules (reverse ID order)
	sort.Slice(visibleRules, func(i, j int) bool {
		// If both have same satisfaction status, sort by ID descending
		if visibleRules[i].IsSatisfied == visibleRules[j].IsSatisfied {
			return visibleRules[i].ID > visibleRules[j].ID
		}
		// Unsatisfied rules come first
		return !visibleRules[i].IsSatisfied && visibleRules[j].IsSatisfied
	})

	return visibleRules
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>The Ultimate Password Game</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <link rel="stylesheet" href="/style.css">
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 The Password Game*</h1>
        </div>
        
        <div class="input-section">
            <div class="input-wrapper">
                <input type="text" 
                       class="password-input" 
                       placeholder="insert here..."
                       hx-post="/validate"
                       hx-target="#rules-container"
                       hx-trigger="input"
                       hx-include="this"
                       name="password"
                       autocomplete="off"
                       value="{{.Password}}"
                       id="password-input"
                       hx-headers='{"X-Prev-Max-Visible": "1"}'>
                <div class="char-count" id="char-count">0</div>
            </div>
        </div>
        
        <div id="rules-container" class="rules-container">
            {{if .HasPassword}}
                {{range .SortedRules}}
                <div class="rule-item {{if .IsSatisfied}}satisfied{{end}}">
                    <div class="rule-number">{{.ID}}</div>
                    <div class="rule-content">
                        <div class="rule-text">{{.Description}}</div>
                        {{if not .IsSatisfied}}
                        <div class="rule-hint">{{.Hint}}</div>
                        {{end}}
                    </div>
                    <div class="checkmark">✓</div>
                </div>
                {{end}}
            {{else}}
            <div style="text-align: center; color: #666; font-style: italic; padding: 20px;">
                Start typing to see the rules appear...
            </div>
            {{end}}
        </div>
    </div>

    <script>
        // Track max visible rule and previous states for animations
        let currentMaxVisible = 1;
        let satisfiedStates = {};
        
        // Update character count on input
        document.addEventListener('DOMContentLoaded', function() {
            const passwordInput = document.querySelector('.password-input');
            const charCount = document.getElementById('char-count');
            
            function updateCharCount() {
                charCount.textContent = passwordInput.value.length;
            }
            
            // Update on page load
            updateCharCount();
            
            // Update on input
            passwordInput.addEventListener('input', updateCharCount);
            
            // Update headers before HTMX request
            passwordInput.addEventListener('htmx:configRequest', function(evt) {
                evt.detail.headers['X-Prev-Max-Visible'] = currentMaxVisible.toString();
                evt.detail.headers['X-Satisfied-States'] = JSON.stringify(satisfiedStates);
            });
            
            // Update current max visible and satisfied states after response
            passwordInput.addEventListener('htmx:afterRequest', function(evt) {
                const newMaxVisible = evt.detail.xhr.getResponseHeader('X-Max-Visible');
                if (newMaxVisible) {
                    currentMaxVisible = parseInt(newMaxVisible);
                }
                
                const newSatisfiedStates = evt.detail.xhr.getResponseHeader('X-Satisfied-States');
                if (newSatisfiedStates) {
                    satisfiedStates = JSON.parse(newSatisfiedStates);
                }
            });
        });
    </script>
</body>
</html>`

const rulesPartialTemplate = `{{if .HasPassword}}
    {{range .SortedRules}}
    <div class="rule-item {{if .IsSatisfied}}satisfied{{end}} {{if .NewlyRevealed}}newly-revealed{{end}} {{if .NewlySatisfied}}newly-satisfied{{end}}">
        <div class="rule-number">{{.ID}}</div>
        <div class="rule-content">
            <div class="rule-text">{{.Description}}</div>
            {{if not .IsSatisfied}}
            <div class="rule-hint">{{.Hint}}</div>
            {{end}}
        </div>
        <div class="checkmark">✓</div>
    </div>
    {{end}}
{{else}}
<div style="text-align: center; color: #666; font-style: italic; padding: 20px;">
    Start typing to see the rules appear...
</div>
{{end}}`

type TemplateData struct {
	Password           string
	Rules              []Rule
	SortedRules        []Rule
	SatisfiedCount     int
	ProgressPercentage float64
	AllSatisfied       bool
	MaxVisibleRule     int
	HasPassword        bool
	PrevMaxVisible     int
}

func getMaxVisibleRule(rules []Rule) int {
	if len(rules) == 0 {
		return 0
	}

	// Show rule 1 if there's any password input
	maxVisible := 1

	// Show next rule if current is satisfied
	for i, rule := range rules {
		if rule.IsSatisfied {
			if i+2 <= len(rules) {
				maxVisible = i + 2 // Show next rule (1-indexed)
			} else {
				maxVisible = len(rules) // All rules visible
			}
		} else {
			break // Stop at first unsatisfied rule
		}
	}

	return maxVisible
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ruleSet := NewRuleSet()
	satisfiedCount := ruleSet.GetSatisfiedCount()

	data := TemplateData{
		Password:           "",
		Rules:              ruleSet.Rules,
		SortedRules:        []Rule{}, // Empty for initial load
		SatisfiedCount:     satisfiedCount,
		ProgressPercentage: 0,
		AllSatisfied:       false,
		HasPassword:        false,
		MaxVisibleRule:     1,
	}

	tmpl := template.Must(template.New("main").Parse(htmlTemplate))
	tmpl.Execute(w, data)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	password := r.FormValue("password")
	prevMaxVisible := 1

	// Try to get previous max visible from header (sent by HTMX)
	if prev := r.Header.Get("X-Prev-Max-Visible"); prev != "" {
		if val, err := strconv.Atoi(prev); err == nil {
			prevMaxVisible = val
		}
	}

	// Get previous satisfied states
	var previousStates []bool
	if states := r.Header.Get("X-Satisfied-States"); states != "" {
		// Parse the JSON map and convert to slice
		stateMap := make(map[string]bool)
		if err := json.Unmarshal([]byte(states), &stateMap); err == nil {
			previousStates = make([]bool, 20) // Assuming 20 rules
			for i := 0; i < 20; i++ {
				previousStates[i] = stateMap[strconv.Itoa(i)]
			}
		}
	}

	ruleSet := NewRuleSet()
	ruleSet.ValidatePassword(password, previousStates)

	satisfiedCount := ruleSet.GetSatisfiedCount()
	rulesLen := len(ruleSet.Rules)
	progressPercentage := (float64(satisfiedCount) / float64(rulesLen)) * 100
	allSatisfied := satisfiedCount == rulesLen
	maxVisible := getMaxVisibleRule(ruleSet.Rules)

	// Get sorted visible rules
	sortedRules := ruleSet.GetSortedVisibleRules(maxVisible)

	// Mark newly revealed rules
	for i := range sortedRules {
		ruleIndex := sortedRules[i].ID - 1 // Convert to 0-based index
		if sortedRules[i].ID > prevMaxVisible && sortedRules[i].ID <= maxVisible && !sortedRules[i].IsSatisfied {
			sortedRules[i].NewlyRevealed = true
		}
		// Update newly satisfied from original rules
		if ruleIndex < len(ruleSet.Rules) {
			sortedRules[i].NewlySatisfied = ruleSet.Rules[ruleIndex].NewlySatisfied
		}
	}

	data := TemplateData{
		Password:           password,
		Rules:              ruleSet.Rules,
		SortedRules:        sortedRules,
		SatisfiedCount:     satisfiedCount,
		ProgressPercentage: progressPercentage,
		AllSatisfied:       allSatisfied,
		HasPassword:        len(password) > 0,
		MaxVisibleRule:     maxVisible,
		PrevMaxVisible:     prevMaxVisible,
	}

	// Send the new max visible and satisfied states back to client
	w.Header().Set("X-Max-Visible", strconv.Itoa(maxVisible))

	// Convert satisfied states to map for JSON
	stateMap := make(map[string]bool)
	for i, rule := range ruleSet.Rules {
		stateMap[strconv.Itoa(i)] = rule.IsSatisfied
	}
	if statesJSON, err := json.Marshal(stateMap); err == nil {
		w.Header().Set("X-Satisfied-States", string(statesJSON))
	}

	// Return just the rules partial for HTMX
	tmpl := template.Must(template.New("rules").Parse(rulesPartialTemplate))
	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/validate", handleValidate)
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})

	log.Println("🚀 Password Game server starting on :8080")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
