package component

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	database "passgame/Database"
	"passgame/rules" // Unified rules package
)

// Template functions
var funcMap = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"subtract": func(a, b int) int {
		return a - b
	},
}

// Global template variable - parse all templates at startup
var tmpl = template.Must(template.New("").Funcs(funcMap).ParseFiles(
	"Frontend/display.html",
	"Frontend/user-modal.html",
))

type PageData struct {
	Password string
	Rules    []rules.Rule
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
var UserSessions = make(map[string]*UserSession)

const rulesPartialTemplate = `{{range $index, $rule := .SortedRules}}
<div class="rule-item {{if .IsSatisfied}}satisfied{{end}} {{if .NewlyRevealed}}newly-revealed{{end}} {{if .NewlySatisfied}}newly-satisfied{{end}}" data-rule-id="{{.ID}}">
    <div class="rule-content">
        <div class="rule-text">{{.Description}}</div>
        
        {{- if eq .ID 14 -}}
        <div class="captcha-container">
            <button type="button" class="update-password-btn" onclick="showRule14Popup({{.ID}})">Update</button>
        </div>
        <div id="rule14-popup-{{.ID}}" class="modal-overlay" style="display:none;z-index:10000;">
            <div class="modal-container" style="text-align:center;">
                <div class="modal-header">
                    <h2>Update Password</h2>
                    <p>Click the button below to reveal your password.</p>
                </div>
                <button type="button" class="btn" onclick="revealRule14Password({{.ID}})">Reveal Password</button>
                <button type="button" class="btn btn-secondary" onclick="hideRule14Popup({{.ID}})">Cancel</button>
            </div>
        </div>
        <div id="rule14-password-{{.ID}}" class="rule14-password" style="display:none;"></div>
        {{- end -}}

        {{if .HasCaptcha}}
        {{- if eq .ID 15 -}}
        <div class="captcha-container">
            <img src="/captcha.png" alt="Captcha" class="captcha-image" id="captcha-{{.ID}}">
            <button type="button" class="refresh-captcha-btn" onclick="refreshCaptcha({{.ID}})">🔄</button>
        </div>
        {{- else if eq .ID 17 -}}
        <div class="qrcode-container">
            <img src="/qrcode.png" alt="QR Code" class="qrcode-image" id="qrcode-{{.ID}}">
            <button type="button" class="refresh-qrcode-btn" onclick="refreshQRCode({{.ID}})">🔄</button>
        </div>
        {{- else if eq .ID 18 -}}
        <div class="color-container">
            <img src="/color.png" alt="Color" class="color-image" id="color-{{.ID}}">
            <button type="button" class="refresh-color-btn" onclick="refreshColor({{.ID}})">🔄</button>
        </div>
        {{- else if eq .ID 19 -}}
        <div class="chess-container">
            <img src="/chess.png" alt="Chess Board" class="chess-image" id="chess-{{.ID}}">
            <button type="button" class="refresh-chess-btn" onclick="refreshChess({{.ID}})">🔄</button>
        </div>
        {{- end -}}
        {{end}}
        
        {{- if eq .ID 20 -}}
        <div class="rule20-progress-container">
            <div class="rule20-progress-bar-bg">
                <div class="rule20-progress-bar" id="rule20-progress-bar-{{.ID}}" style="width:0%"></div>
            </div>
            <div class="rule20-progress-label" id="rule20-progress-label-{{.ID}}">0/3 🏋️</div>
        </div>
        {{- else if eq .ID 22 -}}
        <div class="rule22-pdf-link">
            <a href="#" id="rule22-pdf-link" style="color:blue;text-decoration:underline;cursor:pointer;">pdf file</a>
        </div>
        {{- else if eq .ID 23 -}}
        <div class="watch-ad-container" id="watch-ad-container-{{.ID}}">
            <button id="watch-ad-btn-23" class="btn-primary" onclick="return showAdModal();">Watch Ad to Unlock</button>
        </div>
        <div class="rule23-reveal" style="display: none;"></div>
        {{- end -}}
        
        {{if and (not .IsSatisfied) $.ShowHints}}
        <div class="rule-hint">{{.Hint}}</div>
        {{end}}
    </div>
    <div class="checkmark">✓</div>
</div>
{{end}}`

type TemplateData struct {
	Password           string
	Rules              []rules.Rule
	SortedRules        []rules.Rule
	SatisfiedCount     int
	ProgressPercentage float64
	AllSatisfied       bool
	HasPassword        bool
	RuleChanges        RuleChangeAnalysis
	Title              string
	UserSession        *UserSession
	Difficulties       map[string]DifficultyConfig
	ShowHints          bool
}

func analyzeRuleChanges(currentRules []rules.Rule, previousSatisfied, previousVisible []bool) RuleChangeAnalysis {
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

	session, exists := UserSessions[cookie.Value]
	if !exists {
		return nil
	}

	return session
}

// HandleRegisterUser handles user registration
func HandleRegisterUser(w http.ResponseWriter, r *http.Request) {
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

	// Reset cybersecurity rules for the new session
	rules.ResetCyberSecurityRules()

	UserSessions[sessionID] = userSession

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user_session",
		Value:    sessionID,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
	})

	// Return success response (you might want to redirect or return JSON)
	w.WriteHeader(http.StatusOK)
}

// HandlePasswordGame handles the main password game page
func HandlePasswordGame(w http.ResponseWriter, r *http.Request) {
	// Check if this is a test session request
	if r.URL.Query().Get("test_session") == "true" {
		difficulty := r.URL.Query().Get("difficulty")
		if difficulty == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// This is a test session, create a temporary session
		testUser := &UserSession{
			UserID:     -1, // Negative ID indicates test session
			Username:   "Test User",
			Difficulty: difficulty,
			StartTime:  time.Now(),
			MaxRule:    0,
		}

		// Create a temporary session ID for the test session
		sessionID := "test_" + fmt.Sprint(time.Now().UnixNano())

		// Reset cybersecurity rules for the test session
		rules.ResetCyberSecurityRules()

		UserSessions[sessionID] = testUser

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "user_session",
			Value:    sessionID,
			HttpOnly: true,
			Path:     "/",
			MaxAge:   60 * 60, // 1 hour
		})

		// Redirect to the game
		http.Redirect(w, r, "/display", http.StatusSeeOther)
		return
	}

	// Check if user has a session
	userSession := getUserSession(r)

	if userSession == nil {
		// Show registration modal by executing display.html template with no user session
		data := TemplateData{
			Title:       "The Ultimate Password Game",
			UserSession: nil, // This will trigger the modal to show
		}

		err := tmpl.ExecuteTemplate(w, "display.html", data)
		if err != nil {
			log.Printf("Error executing display template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	ruleSet := rules.NewRuleSet(userSession.Difficulty)

	// Show rule 1 by default (even with empty password)
	ruleSet.Rules[0].IsVisible = true

	satisfiedCount := rules.GetSatisfiedCount(ruleSet)
	sortedRules := rules.GetSortedVisibleRules(ruleSet)

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
		ShowHints:          Config.ShowHints,
	}

	// Execute the display.html template with data
	err := tmpl.ExecuteTemplate(w, "display.html", data)
	if err != nil {
		log.Printf("Error executing display template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleUserModal handles user modal requests
func HandleUserModal(w http.ResponseWriter, r *http.Request) {
	difficulties, err := LoadDifficulties()
	if err != nil {
		log.Printf("Warning: Could not load difficulties: %v", err)
	}

	data := TemplateData{
		Title:        "User Registration",
		Difficulties: difficulties,
	}

	err = tmpl.ExecuteTemplate(w, "user-modal.html", data)
	if err != nil {
		log.Printf("Error executing user-modal template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleValidate handles password validation
func HandleValidate(w http.ResponseWriter, r *http.Request) {
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

	// Create rule set based on user's difficulty
	ruleSet := rules.NewRuleSet(userSession.Difficulty)

	// Get previous satisfied states
	var previousSatisfiedStates []bool
	if states := r.Header.Get("X-Satisfied-States"); states != "" {
		stateMap := make(map[string]bool)
		if err := json.Unmarshal([]byte(states), &stateMap); err == nil {
			previousSatisfiedStates = make([]bool, len(ruleSet.Rules))
			for i := 0; i < len(ruleSet.Rules); i++ {
				previousSatisfiedStates[i] = stateMap[strconv.Itoa(ruleSet.Rules[i].ID)] // Use actual rule ID
			}
		}
	}

	// Get previous visible states
	var previousVisibleStates []bool
	if states := r.Header.Get("X-Visible-States"); states != "" {
		stateMap := make(map[string]bool)
		if err := json.Unmarshal([]byte(states), &stateMap); err == nil {
			previousVisibleStates = make([]bool, len(ruleSet.Rules))
			for i := 0; i < len(ruleSet.Rules); i++ {
				previousVisibleStates[i] = stateMap[strconv.Itoa(ruleSet.Rules[i].ID)] // Use actual rule ID
			}
		}
	}

	rules.ValidatePassword(ruleSet, password, previousSatisfiedStates, previousVisibleStates)

	// Track if we need to update the database
	shouldUpdateDB := false
	highestNewlySatisfiedRule := 0

	// Check for newly satisfied rules
	for _, rule := range ruleSet.Rules {
		if rule.NewlySatisfied {
			shouldUpdateDB = true
			if rule.ID > highestNewlySatisfiedRule {
				highestNewlySatisfiedRule = rule.ID
			}
			log.Printf("✅ Rule %d newly satisfied for user %s", rule.ID, userSession.Username)
		}
	}

	// Only update database if there are newly satisfied rules AND it's a higher rule than previously reached
	if shouldUpdateDB && highestNewlySatisfiedRule > userSession.MaxRule {
		timeSpent := int(time.Since(userSession.StartTime).Seconds())

		// Update max rule reached in session
		userSession.MaxRule = highestNewlySatisfiedRule

		// Update database
		err := database.UpdateUserProgress(userSession.UserID, highestNewlySatisfiedRule, timeSpent)
		if err != nil {
			log.Printf("Error updating user progress for rule %d: %v", highestNewlySatisfiedRule, err)
		} else {
			log.Printf("📈 Database updated for user %s: Rule %d satisfied in %ds",
				userSession.Username, highestNewlySatisfiedRule, timeSpent)
		}
	}

	// Check if all rules are satisfied (game completed)
	satisfiedCount := rules.GetSatisfiedCount(ruleSet)
	rulesLen := len(ruleSet.Rules)
	if satisfiedCount == rulesLen && !userSession.IsCompleted {
		userSession.IsCompleted = true
		timeSpent := int(time.Since(userSession.StartTime).Seconds())

		err := database.UpdateUserProgress(userSession.UserID, rulesLen, timeSpent) // Use actual rule count
		if err != nil {
			log.Printf("Error updating completion: %v", err)
		} else {
			log.Printf("🎉 Game completed by user %s in %d seconds!", userSession.Username, timeSpent)
		}
	}

	// Analyze what changed
	ruleChanges := analyzeRuleChanges(ruleSet.Rules, previousSatisfiedStates, previousVisibleStates)

	progressPercentage := (float64(satisfiedCount) / float64(rulesLen)) * 100
	allSatisfied := satisfiedCount == rulesLen

	// Get sorted visible rules
	sortedRules := rules.GetSortedVisibleRules(ruleSet)

	data := TemplateData{
		Password:           password,
		Rules:              ruleSet.Rules,
		SortedRules:        sortedRules,
		SatisfiedCount:     satisfiedCount,
		ProgressPercentage: progressPercentage,
		AllSatisfied:       allSatisfied,
		HasPassword:        len(password) > 0,
		RuleChanges:        ruleChanges,
		ShowHints:          Config.ShowHints,
		UserSession:        userSession,
	}

	// Send the satisfied and visible states back to client
	satisfiedStateMap := make(map[string]bool)
	visibleStateMap := make(map[string]bool)
	for _, rule := range ruleSet.Rules {
		satisfiedStateMap[strconv.Itoa(rule.ID)] = rule.IsSatisfied
		visibleStateMap[strconv.Itoa(rule.ID)] = rule.IsVisible
	}

	if statesJSON, err := json.Marshal(satisfiedStateMap); err == nil {
		w.Header().Set("X-Satisfied-States", string(statesJSON))
	}

	if visibleJSON, err := json.Marshal(visibleStateMap); err == nil {
		w.Header().Set("X-Visible-States", string(visibleJSON))
	}

	// Return just the rules partial for HTMX
	ruleTmpl := template.Must(template.New("rules").Funcs(funcMap).Parse(rulesPartialTemplate))
	ruleTmpl.Execute(w, data)
}
