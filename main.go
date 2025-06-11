package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"passgame/rules/basic" // Basic rules package
)

type PageData struct {
	Password string
	Rules    []basic.Rule
}

// RuleChangeAnalysis tracks what changed between validations
type RuleChangeAnalysis struct {
	HasChanges       bool
	NewlySatisfied   []int
	NewlyUnsatisfied []int
	NewlyVisible     []int
	NewlyHidden      []int
}

const rulesPartialTemplate = `{{range .SortedRules}}
<div class="rule-item {{if .IsSatisfied}}satisfied{{end}} {{if .NewlyRevealed}}newly-revealed{{end}} {{if .NewlySatisfied}}newly-satisfied{{end}}" data-rule-id="{{.ID}}">
    <div class="rule-number">{{.ID}}</div>
    <div class="rule-content">
        <div class="rule-text">{{.Description}}</div>
        {{if not .IsSatisfied}}
        <div class="rule-hint">{{.Hint}}</div>
        {{end}}
    </div>
    <div class="checkmark">✓</div>
</div>
{{end}}`

type TemplateData struct {
	Password           string
	Rules              []basic.Rule
	SortedRules        []basic.Rule
	SatisfiedCount     int
	ProgressPercentage float64
	AllSatisfied       bool
	HasPassword        bool
	RuleChanges        RuleChangeAnalysis
}

func analyzeRuleChanges(currentRules []basic.Rule, previousSatisfied, previousVisible []bool) RuleChangeAnalysis {
	analysis := RuleChangeAnalysis{
		NewlySatisfied:   make([]int, 0),
		NewlyUnsatisfied: make([]int, 0),
		NewlyVisible:     make([]int, 0),
		NewlyHidden:      make([]int, 0),
	}

	for i, rule := range currentRules {
		// Check satisfaction changes
		if i < len(previousSatisfied) {
			wasStatisfied := previousSatisfied[i]
			isStatisfied := rule.IsSatisfied

			if !wasStatisfied && isStatisfied {
				analysis.NewlySatisfied = append(analysis.NewlySatisfied, rule.ID)
				analysis.HasChanges = true
			} else if wasStatisfied && !isStatisfied {
				analysis.NewlyUnsatisfied = append(analysis.NewlyUnsatisfied, rule.ID)
				analysis.HasChanges = true
			}
		}

		// Check visibility changes
		if i < len(previousVisible) {
			wasVisible := previousVisible[i]
			isVisible := rule.IsVisible

			if !wasVisible && isVisible {
				analysis.NewlyVisible = append(analysis.NewlyVisible, rule.ID)
				analysis.HasChanges = true
			} else if wasVisible && !isVisible {
				analysis.NewlyHidden = append(analysis.NewlyHidden, rule.ID)
				analysis.HasChanges = true
			}
		} else if rule.IsVisible {
			// New rule that's visible
			analysis.NewlyVisible = append(analysis.NewlyVisible, rule.ID)
			analysis.HasChanges = true
		}
	}

	return analysis
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	ruleSet := basic.NewRuleSet()

	// Show rule 1 by default (even with empty password)
	ruleSet.Rules[0].IsVisible = true

	satisfiedCount := basic.GetSatisfiedCount(ruleSet)
	sortedRules := basic.GetSortedVisibleRules(ruleSet)

	data := TemplateData{
		Password:           "",
		Rules:              ruleSet.Rules,
		SortedRules:        sortedRules,
		SatisfiedCount:     satisfiedCount,
		ProgressPercentage: 0,
		AllSatisfied:       false,
		HasPassword:        false,
	}

	// Parse and execute the display.html template from Frontend directory
	tmpl, err := template.ParseFiles("Frontend/display.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	password := r.FormValue("password")

	// Get previous satisfied states
	var previousSatisfiedStates []bool
	if states := r.Header.Get("X-Satisfied-States"); states != "" {
		stateMap := make(map[string]bool)
		if err := json.Unmarshal([]byte(states), &stateMap); err == nil {
			previousSatisfiedStates = make([]bool, 20) // Assuming 20 rules
			for i := 0; i < 20; i++ {
				previousSatisfiedStates[i] = stateMap[strconv.Itoa(i)]
			}
		}
	}

	// Get previous visible states
	var previousVisibleStates []bool
	if states := r.Header.Get("X-Visible-States"); states != "" {
		stateMap := make(map[string]bool)
		if err := json.Unmarshal([]byte(states), &stateMap); err == nil {
			previousVisibleStates = make([]bool, 20) // Assuming 20 rules
			for i := 0; i < 20; i++ {
				previousVisibleStates[i] = stateMap[strconv.Itoa(i)]
			}
		}
	}

	ruleSet := basic.NewRuleSet()
	basic.ValidatePassword(ruleSet, password, previousSatisfiedStates, previousVisibleStates)

	// Analyze what changed
	ruleChanges := analyzeRuleChanges(ruleSet.Rules, previousSatisfiedStates, previousVisibleStates)

	satisfiedCount := basic.GetSatisfiedCount(ruleSet)
	rulesLen := len(ruleSet.Rules)
	progressPercentage := (float64(satisfiedCount) / float64(rulesLen)) * 100
	allSatisfied := satisfiedCount == rulesLen

	// Get sorted visible rules
	sortedRules := basic.GetSortedVisibleRules(ruleSet)

	data := TemplateData{
		Password:           password,
		Rules:              ruleSet.Rules,
		SortedRules:        sortedRules,
		SatisfiedCount:     satisfiedCount,
		ProgressPercentage: progressPercentage,
		AllSatisfied:       allSatisfied,
		HasPassword:        len(password) > 0,
		RuleChanges:        ruleChanges,
	}

	// Send the satisfied and visible states back to client
	satisfiedStateMap := make(map[string]bool)
	visibleStateMap := make(map[string]bool)
	for i, rule := range ruleSet.Rules {
		satisfiedStateMap[strconv.Itoa(i)] = rule.IsSatisfied
		visibleStateMap[strconv.Itoa(i)] = rule.IsVisible
	}

	if statesJSON, err := json.Marshal(satisfiedStateMap); err == nil {
		w.Header().Set("X-Satisfied-States", string(statesJSON))
	}

	if visibleJSON, err := json.Marshal(visibleStateMap); err == nil {
		w.Header().Set("X-Visible-States", string(visibleJSON))
	}

	// Return just the rules partial for HTMX
	tmpl := template.Must(template.New("rules").Parse(rulesPartialTemplate))
	tmpl.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/validate", handleValidate)

	// Serve static files from Frontend directory
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "Frontend/style.css")
	})

	log.Println("🚀 Password Game server starting on :8080")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
