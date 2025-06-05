package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Rule struct {
	ID          int
	Description string
	Validator   func(string) bool
	IsSatisfied bool
	Hint        string
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

func (rs *RuleSet) ValidatePassword(password string) {
	for i := range rs.Rules {
		rs.Rules[i].IsSatisfied = rs.Rules[i].Validator(password)
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
            <h1>🔐 Password Game</h1>
        </div>
        
        <div class="input-section">
            <input type="text" 
                   class="password-input" 
                   placeholder="Start typing your password..."
                   hx-post="/validate"
                   hx-target="#rules-container"
                   hx-trigger="keyup changed delay:300ms"
                   hx-include="this"
                   name="password"
                   autocomplete="off"
                   value="{{.Password}}">
                   
            <div class="progress-bar">
                <div class="progress-fill" style="width: {{.ProgressPercentage}}%"></div>
            </div>

        </div>
        
            
            <div id="rules-container" class="rules-container">
                {{if .HasPassword}}
                    {{range $index, $rule := .Rules}}
                        {{if or (lt $index $.MaxVisibleRule) ($rule.IsSatisfied)}}
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
                    {{end}}
                {{else}}
                <div style="text-align: center; color: #666; font-style: italic; padding: 20px;">
                    Enter Password
                </div>
                {{end}}
            </div>
        </div>
    </div>
</body>
</html>`

const rulesPartialTemplate = `{{if .HasPassword}}
    {{range $index, $rule := .Rules}}
        {{if or (lt $index $.MaxVisibleRule) ($rule.IsSatisfied)}}
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
    {{end}}
{{else}}
<div style="text-align: center; color: #666; font-style: italic; padding: 20px;">
    Start typing to see the rules appear...
</div>
{{end}}`

type TemplateData struct {
	Password           string
	Rules              []Rule
	SatisfiedCount     int
	ProgressPercentage float64
	AllSatisfied       bool
	MaxVisibleRule     int
	HasPassword        bool
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
	ruleSet := NewRuleSet()
	ruleSet.ValidatePassword(password)

	satisfiedCount := ruleSet.GetSatisfiedCount()
	rulesLen := len(ruleSet.Rules)
	progressPercentage := (float64(satisfiedCount) / float64(rulesLen)) * 100
	allSatisfied := satisfiedCount == rulesLen
	maxVisible := getMaxVisibleRule(ruleSet.Rules)

	data := TemplateData{
		Password:           password,
		Rules:              ruleSet.Rules,
		SatisfiedCount:     satisfiedCount,
		ProgressPercentage: progressPercentage,
		AllSatisfied:       allSatisfied,
		HasPassword:        len(password) > 0,
		MaxVisibleRule:     maxVisible,
	}

	// Return just the rules partial for HTMX l
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
