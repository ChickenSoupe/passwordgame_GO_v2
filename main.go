package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	database "passgame/Database"
	"passgame/component"
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

// UserSession tracks user session data
type UserSession struct {
	UserID      int64     `json:"user_id"`
	Username    string    `json:"username"`
	Difficulty  string    `json:"difficulty"`
	StartTime   time.Time `json:"start_time"`
	MaxRule     int       `json:"max_rule"`
	IsCompleted bool      `json:"is_completed"`
}

// Global session storage (in production, use Redis or similar)
var userSessions = make(map[string]*UserSession)

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
	Title              string
	UserSession        *UserSession
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

// Generate a simple session ID (in production, use crypto/rand)
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// Get user session from cookie
func getUserSession(r *http.Request) *UserSession {
	cookie, err := r.Cookie("user_session")
	if err != nil {
		return nil
	}

	session, exists := userSessions[cookie.Value]
	if !exists {
		return nil
	}

	return session
}

// Handle user registration
func handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	difficulty := r.FormValue("difficulty")

	// Validate input
	if len(username) < 3 || len(username) > 20 {
		http.Error(w, `<div class="error-message">Username must be between 3-20 characters</div>`, http.StatusBadRequest)
		return
	}

	if difficulty == "" {
		http.Error(w, `<div class="error-message">Please select a difficulty level</div>`, http.StatusBadRequest)
		return
	}

	// Check if username exists
	exists, err := database.CheckUsernameExists(username)
	if err != nil {
		log.Printf("Error checking username: %v", err)
		http.Error(w, `<div class="error-message">Database error occurred</div>`, http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, `<div class="error-message">Username already exists. Please choose another.</div>`, http.StatusBadRequest)
		return
	}

	// Insert user into database
	userID, err := database.InsertUser(username, difficulty)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		http.Error(w, `<div class="error-message">Failed to create user account</div>`, http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID := generateSessionID()
	userSession := &UserSession{
		UserID:     userID,
		Username:   username,
		Difficulty: difficulty,
		StartTime:  time.Now(),
		MaxRule:    0,
	}

	userSessions[sessionID] = userSession

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_session",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
	})

	// Return success response that closes modal
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<div class="success-message">
			<h3>🎉 Welcome, ` + username + `!</h3>
			<p>Your account has been created successfully.</p>
		</div>
		<script>
			setTimeout(function() {
				document.getElementById('user-modal').style.display = 'none';
				document.querySelector('.password-input').focus();
			}, 1500);
		</script>
	`))
}

// Handle the main password game page
func handlePasswordGame(w http.ResponseWriter, r *http.Request) {
	// Check if user has a session
	userSession := getUserSession(r)
	if userSession == nil {
		// Show registration modal
		http.ServeFile(w, r, "Frontend/display.html")
		return
	}

	ruleSet := basic.NewRuleSet()

	// Show rule 1 by default (even with empty password)
	ruleSet.Rules[0].IsVisible = true

	satisfiedCount := basic.GetSatisfiedCount(ruleSet)
	sortedRules := basic.GetSortedVisibleRules(ruleSet)

	data := TemplateData{
		Title:              "The Ultimate Password Game",
		Password:           "",
		Rules:              ruleSet.Rules,
		SortedRules:        sortedRules,
		SatisfiedCount:     satisfiedCount,
		ProgressPercentage: 0,
		AllSatisfied:       false,
		HasPassword:        false,
		UserSession:        userSession,
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

	// Check user session
	userSession := getUserSession(r)
	if userSession == nil {
		http.Error(w, "Session expired", http.StatusUnauthorized)
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

	// Update user progress
	maxRuleReached := 0
	for _, rule := range ruleSet.Rules {
		if rule.IsVisible {
			if rule.ID > maxRuleReached {
				maxRuleReached = rule.ID
			}
		}
	}

	// Update session and database if new max rule reached
	if maxRuleReached > userSession.MaxRule {
		userSession.MaxRule = maxRuleReached
		timeSpent := int(time.Since(userSession.StartTime).Seconds())

		err := database.UpdateUserProgress(userSession.UserID, maxRuleReached, timeSpent)
		if err != nil {
			log.Printf("Error updating user progress: %v", err)
		}
	}

	// Check if all rules are satisfied (game completed)
	satisfiedCount := basic.GetSatisfiedCount(ruleSet)
	rulesLen := len(ruleSet.Rules)
	if satisfiedCount == rulesLen && !userSession.IsCompleted {
		userSession.IsCompleted = true
		timeSpent := int(time.Since(userSession.StartTime).Seconds())

		err := database.UpdateUserProgress(userSession.UserID, 20, timeSpent) // 20 = all rules completed
		if err != nil {
			log.Printf("Error updating completion: %v", err)
		}
	}

	// Analyze what changed
	ruleChanges := analyzeRuleChanges(ruleSet.Rules, previousSatisfiedStates, previousVisibleStates)

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
		UserSession:        userSession,
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
	// Initialize database
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Create Database directory if it doesn't exist
	if err := os.MkdirAll("Database", 0755); err != nil {
		log.Printf("Warning: Could not create Database directory: %v", err)
	}

	// Main routes - both root and /display point to the same handler
	http.HandleFunc("/", handlePasswordGame)
	http.HandleFunc("/display", handlePasswordGame)
	http.HandleFunc("/validate", handleValidate)
	http.HandleFunc("/register-user", handleRegisterUser)
	http.HandleFunc("/leaderboard", component.HandleLeaderboard)

	// Serve static files from Frontend directory
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "Frontend/style.css")
	})

	http.HandleFunc("/flip-animations.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "Frontend/flip-animations.js")
	})

	http.HandleFunc("/user-modal.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "Frontend/user-modal.html")
	})

	log.Println("🚀 Password Game server starting on :8080")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Println("🎮 Password Game: http://localhost:8080/display")
	log.Println("🏆 Leaderboard: http://localhost:8080/leaderboard")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
