package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}

// DifficultyConfig represents the configuration for a difficulty level
type DifficultyConfig struct {
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// User represents a user in the database
type User struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Difficulty  string    `json:"difficulty"`
	RuleReached int       `json:"rule_reached"`
	TimeSpent   int       `json:"time_spent"` // in seconds
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SortConfig holds sorting configuration
type SortConfig struct {
	Column string
	Order  string
}

// Valid sort columns mapping
var validSortColumns = map[string]string{
	"rule":       "rule_reached",
	"time":       "time_spent",
	"difficulty": "difficulty",
	"joined":     "created_at",
	"username":   "username",
}

// LoadDifficulties loads difficulty configurations from JSON file
func LoadDifficulties() (map[string]DifficultyConfig, error) {
	data, err := ioutil.ReadFile("config/difficulties.json")
	if err != nil {
		log.Printf("Error reading difficulties.json: %v", err)
		return getDefaultDifficulties(), err
	}

	var difficulties map[string]DifficultyConfig
	if err := json.Unmarshal(data, &difficulties); err != nil {
		log.Printf("Error parsing difficulties.json: %v", err)
		return getDefaultDifficulties(), err
	}

	return difficulties, nil
}

// getDefaultDifficulties returns hardcoded default difficulties as fallback
func getDefaultDifficulties() map[string]DifficultyConfig {
	return map[string]DifficultyConfig{
		"basic": {
			Name:        "Basic",
			Icon:        "🟢",
			Color:       "#4CAF50",
			Description: "Standard rules",
		},
		"intermediate": {
			Name:        "Intermediate",
			Icon:        "🟡",
			Color:       "#FF9800",
			Description: "More challenging",
		},
		"hard": {
			Name:        "Hard",
			Icon:        "🔴",
			Color:       "#F44336",
			Description: "Expert level",
		},
		"expert": {
			Name:        "Expert",
			Icon:        "🟣",
			Color:       "#9C27B0",
			Description: "Master level",
		},
		"fun": {
			Name:        "Fun",
			Icon:        "🎉",
			Color:       "#E91E63",
			Description: "Quirky rules",
		},
	}
}

// getDynamicDifficulties gets valid difficulties from the config
func getDynamicDifficulties() []string {
	difficulties, err := LoadDifficulties()
	if err != nil {
		// Fallback to default difficulties if config loading fails
		return []string{"basic", "intermediate", "hard", "expert", "fun"}
	}

	var validDiffs []string
	for key := range difficulties {
		validDiffs = append(validDiffs, key)
	}
	return validDiffs
}

// InitDB initializes the SQLite database with improved schema
func InitDB() error {
	var err error

	// Create the database file in the Database directory
	db, err = sql.Open("sqlite", "Database/user.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create the users table with improved schema
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL COLLATE NOCASE,
		difficulty TEXT NOT NULL CHECK(difficulty IN ('basic', 'intermediate', 'hard', 'expert', 'fun')),
		rule_reached INTEGER DEFAULT 0 CHECK(rule_reached >= 0 AND rule_reached <= 50),
		time_spent INTEGER DEFAULT 0 CHECK(time_spent >= 0),
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Create optimized indexes
	CREATE INDEX IF NOT EXISTS idx_username ON users(username COLLATE NOCASE);
	CREATE INDEX IF NOT EXISTS idx_leaderboard ON users(rule_reached DESC, time_spent ASC);
	CREATE INDEX IF NOT EXISTS idx_difficulty_progress ON users(difficulty, rule_reached DESC);
	CREATE INDEX IF NOT EXISTS idx_created_at ON users(created_at);
	
	-- Create trigger to automatically update updated_at
	CREATE TRIGGER IF NOT EXISTS update_users_updated_at 
		AFTER UPDATE ON users
		FOR EACH ROW
		BEGIN
			UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END;
	`

	if _, err = db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create table and indexes: %v", err)
	}

	log.Println("✅ Database initialized successfully with optimized schema")
	return nil
}

// CloseDB closes the database connection gracefully
func CloseDB() error {
	if db != nil {
		log.Println("🔌 Closing database connection...")
		return db.Close()
	}
	return nil
}

// CheckUsernameExists checks if a username already exists (case-insensitive)
func CheckUsernameExists(username string) (bool, error) {
	if strings.TrimSpace(username) == "" {
		return false, fmt.Errorf("username cannot be empty")
	}

	var count int
	query := "SELECT COUNT(*) FROM users WHERE username = ? COLLATE NOCASE"

	err := db.QueryRow(query, strings.TrimSpace(username)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %v", err)
	}

	return count > 0, nil
}

// ValidateDifficulty checks if the difficulty is valid
func ValidateDifficulty(difficulty string) bool {
	if difficulty == "all" {
		return true
	}

	diffs, err := LoadDifficulties()
	if err != nil {
		return false
	}

	// Check against loaded difficulties (case-insensitive)
	for k := range diffs {
		if strings.EqualFold(difficulty, k) {
			return true
		}
	}
	return false
}

// InsertUser inserts a new user with validation
func InsertUser(username, difficulty string) (int64, error) {
	// Validate inputs
	username = strings.TrimSpace(username)
	difficulty = strings.ToLower(strings.TrimSpace(difficulty))

	if username == "" {
		return 0, fmt.Errorf("username cannot be empty")
	}

	if len(username) > 50 {
		return 0, fmt.Errorf("username too long (max 50 characters)")
	}

	if !ValidateDifficulty(difficulty) {
		validDiffs := getDynamicDifficulties()
		return 0, fmt.Errorf("invalid difficulty: %s (valid: %v)", difficulty, validDiffs)
	}

	// Check if username exists
	exists, err := CheckUsernameExists(username)
	if err != nil {
		return 0, fmt.Errorf("failed to check username: %v", err)
	}
	if exists {
		return 0, fmt.Errorf("username '%s' already exists", username)
	}

	// Insert user
	query := `
		INSERT INTO users (username, difficulty, rule_reached, time_spent, created_at, updated_at)
		VALUES (?, ?, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := db.Exec(query, username, difficulty)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %v", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID: %v", err)
	}

	log.Printf("✅ User created: %s (ID: %d, Difficulty: %s)", username, userID, difficulty)
	return userID, nil
}

// UpdateUserProgress updates user progress with validation
func UpdateUserProgress(userID int64, ruleReached, timeSpent int) error {
	// Validate inputs
	if userID <= 0 {
		return fmt.Errorf("invalid user ID: %d", userID)
	}
	if ruleReached < 0 || ruleReached > 50 {
		return fmt.Errorf("invalid rule reached: %d (must be 0-50)", ruleReached)
	}
	if timeSpent < 0 {
		return fmt.Errorf("invalid time spent: %d (must be >= 0)", timeSpent)
	}

	query := `
		UPDATE users 
		SET rule_reached = ?, time_spent = ?
		WHERE id = ?
	`

	result, err := db.Exec(query, ruleReached, timeSpent, userID)
	if err != nil {
		return fmt.Errorf("failed to update user progress: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with ID: %d", userID)
	}

	log.Printf("📈 Progress updated for user ID %d: Rule %d, Time %ds", userID, ruleReached, timeSpent)
	return nil
}

// GetUser retrieves a user by ID with error handling
func GetUser(userID int64) (*User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}

	query := `
		SELECT id, username, difficulty, rule_reached, time_spent, created_at, updated_at
		FROM users WHERE id = ?
	`

	user := &User{}
	err := db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Difficulty,
		&user.RuleReached,
		&user.TimeSpent,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", userID)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username (case-insensitive)
func GetUserByUsername(username string) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	query := `
		SELECT id, username, difficulty, rule_reached, time_spent, created_at, updated_at
		FROM users WHERE username = ? COLLATE NOCASE
	`

	user := &User{}
	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Difficulty,
		&user.RuleReached,
		&user.TimeSpent,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user '%s' not found", username)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// GetLeaderboard retrieves the top users with default sorting
func GetLeaderboard(limit int) ([]User, error) {
	return GetLeaderboardSorted(limit, "rule", "desc")
}

// GetLeaderboardSorted retrieves users with custom sorting and filtering
func GetLeaderboardSorted(limit int, sortBy, sortOrder string) ([]User, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Prevent excessive queries
	}

	// Validate and normalize sort parameters
	sortConfig := validateSortConfig(sortBy, sortOrder)
	orderBy := buildOrderByClause(sortConfig)

	query := fmt.Sprintf(`
		SELECT id, username, difficulty, rule_reached, time_spent, created_at, updated_at
		FROM users 
		ORDER BY %s
		LIMIT ?
	`, orderBy)

	return executeUserQuery(query, limit)
}

// GetLeaderboardByDifficulty retrieves users filtered by difficulty
func GetLeaderboardByDifficulty(difficulty string, limit int, sortBy, sortOrder string) ([]User, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Validate difficulty
	difficulty = strings.ToLower(strings.TrimSpace(difficulty))
	if !ValidateDifficulty(difficulty) {
		return nil, fmt.Errorf("invalid difficulty: %s", difficulty)
	}

	// Validate and normalize sort parameters
	sortConfig := validateSortConfig(sortBy, sortOrder)
	orderBy := buildOrderByClause(sortConfig)

	query := fmt.Sprintf(`
		SELECT id, username, difficulty, rule_reached, time_spent, created_at, updated_at
		FROM users 
		WHERE difficulty = ?
		ORDER BY %s
		LIMIT ?
	`, orderBy)

	return executeUserQueryWithParam(query, difficulty, limit)
}

// validateSortConfig validates and normalizes sort configuration
func validateSortConfig(sortBy, sortOrder string) SortConfig {
	// Validate sort column
	columnName, valid := validSortColumns[sortBy]
	if !valid {
		columnName = "rule_reached"
		sortBy = "rule"
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return SortConfig{
		Column: columnName,
		Order:  sortOrder,
	}
}

// buildOrderByClause builds the ORDER BY clause based on sort configuration
func buildOrderByClause(config SortConfig) string {
	switch {
	case strings.Contains(config.Column, "rule_reached"):
		if config.Order == "desc" {
			return "rule_reached DESC, time_spent ASC, created_at DESC"
		}
		return "rule_reached ASC, time_spent DESC, created_at DESC"

	case strings.Contains(config.Column, "time_spent"):
		if config.Order == "desc" {
			return "time_spent DESC, rule_reached DESC, created_at DESC"
		}
		return "time_spent ASC, rule_reached DESC, created_at DESC"

	case strings.Contains(config.Column, "difficulty"):
		if config.Order == "desc" {
			return `CASE difficulty 
				WHEN 'expert' THEN 1 
				WHEN 'hard' THEN 2 
				WHEN 'intermediate' THEN 3 
				WHEN 'basic' THEN 4 
				WHEN 'fun' THEN 5 
				ELSE 6 END ASC, rule_reached DESC, time_spent ASC`
		}
		return `CASE difficulty 
			WHEN 'basic' THEN 1 
			WHEN 'intermediate' THEN 2 
			WHEN 'hard' THEN 3 
			WHEN 'expert' THEN 4 
			WHEN 'fun' THEN 5 
			ELSE 6 END ASC, rule_reached DESC, time_spent ASC`

	case strings.Contains(config.Column, "created_at"):
		return fmt.Sprintf("created_at %s, rule_reached DESC, time_spent ASC", strings.ToUpper(config.Order))

	case strings.Contains(config.Column, "username"):
		return fmt.Sprintf("username COLLATE NOCASE %s, rule_reached DESC, time_spent ASC", strings.ToUpper(config.Order))

	default:
		return "rule_reached DESC, time_spent ASC, created_at DESC"
	}
}

// executeUserQuery executes a user query and returns the results
func executeUserQuery(query string, limit int) ([]User, error) {
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	return scanUsers(rows)
}

// executeUserQueryWithParam executes a user query with an additional parameter
func executeUserQueryWithParam(query string, param interface{}, limit int) ([]User, error) {
	rows, err := db.Query(query, param, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	return scanUsers(rows)
}

// scanUsers scans database rows into User structs
func scanUsers(rows *sql.Rows) ([]User, error) {
	var users []User

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Difficulty,
			&user.RuleReached,
			&user.TimeSpent,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return users, nil
}

// GetUserStats returns comprehensive statistics about users
func GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total users
	var totalUsers int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %v", err)
	}
	stats["total_users"] = totalUsers

	if totalUsers == 0 {
		// Return empty stats if no users
		stats["by_difficulty"] = make(map[string]int)
		stats["highest_rule"] = 0
		stats["average_time"] = 0.0
		stats["completion_rates"] = make(map[string]float64)
		return stats, nil
	}

	// Users by difficulty
	diffStats, err := getUsersByDifficulty()
	if err != nil {
		return nil, err
	}
	stats["by_difficulty"] = diffStats

	// Highest rule reached
	var maxRule int
	err = db.QueryRow("SELECT COALESCE(MAX(rule_reached), 0) FROM users").Scan(&maxRule)
	if err != nil {
		return nil, fmt.Errorf("failed to get max rule: %v", err)
	}
	stats["highest_rule"] = maxRule

	// Average time spent (only for users who have played)
	var avgTime float64
	err = db.QueryRow("SELECT COALESCE(AVG(time_spent), 0) FROM users WHERE time_spent > 0").Scan(&avgTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get average time: %v", err)
	}
	stats["average_time"] = avgTime

	// Completion rates by rule
	completionRates, err := getCompletionRates()
	if err != nil {
		return nil, err
	}
	stats["completion_rates"] = completionRates

	return stats, nil
}

// getUsersByDifficulty gets user count by difficulty
func getUsersByDifficulty() (map[string]int, error) {
	diffQuery := `
		SELECT difficulty, COUNT(*) as count 
		FROM users 
		GROUP BY difficulty 
		ORDER BY 
			CASE difficulty 
				WHEN 'basic' THEN 1 
				WHEN 'intermediate' THEN 2 
				WHEN 'hard' THEN 3 
				WHEN 'expert' THEN 4 
				WHEN 'fun' THEN 5 
				ELSE 6 
			END
	`

	rows, err := db.Query(diffQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get difficulty stats: %v", err)
	}
	defer rows.Close()

	diffStats := make(map[string]int)
	for rows.Next() {
		var difficulty string
		var count int
		if err := rows.Scan(&difficulty, &count); err != nil {
			return nil, fmt.Errorf("failed to scan difficulty stats: %v", err)
		}
		diffStats[difficulty] = count
	}

	return diffStats, nil
}

// getCompletionRates calculates completion rates for different rule milestones
func getCompletionRates() (map[string]float64, error) {
	milestones := []int{5, 10, 15, 20}
	rates := make(map[string]float64)

	var totalUsers int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE time_spent > 0").Scan(&totalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total active users: %v", err)
	}

	if totalUsers == 0 {
		for _, milestone := range milestones {
			rates[fmt.Sprintf("rule_%d", milestone)] = 0.0
		}
		return rates, nil
	}

	for _, milestone := range milestones {
		var completedUsers int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE rule_reached >= ?", milestone).Scan(&completedUsers)
		if err != nil {
			return nil, fmt.Errorf("failed to get completion rate for rule %d: %v", milestone, err)
		}

		rate := (float64(completedUsers) / float64(totalUsers)) * 100
		rates[fmt.Sprintf("rule_%d", milestone)] = rate
	}

	return rates, nil
}

// DeleteUser deletes a user from the database with validation
func DeleteUser(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID: %d", userID)
	}

	query := "DELETE FROM users WHERE id = ?"

	result, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user found with ID: %d", userID)
	}

	log.Printf("🗑️ User deleted: ID %d", userID)
	return nil
}

// GetUserCount returns the total number of users
func GetUserCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %v", err)
	}
	return count, nil
}

// GetRecentUsers returns recently joined users
func GetRecentUsers(limit int) ([]User, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	query := `
		SELECT id, username, difficulty, rule_reached, time_spent, created_at, updated_at
		FROM users 
		ORDER BY created_at DESC
		LIMIT ?
	`

	return executeUserQuery(query, limit)
}

// HealthCheck performs a basic database health check
func HealthCheck() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	// Test a simple query
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users LIMIT 1").Scan(&count)
	if err != nil {
		return fmt.Errorf("database query test failed: %v", err)
	}

	return nil
}
