package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

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

// InitDB initializes the SQLite database
func InitDB() error {
	var err error

	// Create the database file in the Database directory
	db, err = sql.Open("sqlite", "Database/user.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Create the users table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		difficulty TEXT NOT NULL,
		rule_reached INTEGER DEFAULT 0,
		time_spent INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_rule_reached ON users(rule_reached);
	CREATE INDEX IF NOT EXISTS idx_difficulty ON users(difficulty);
	`

	if _, err = db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	log.Println("✅ Database initialized successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// CheckUsernameExists checks if a username already exists in the database
func CheckUsernameExists(username string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM users WHERE username = ?"

	err := db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check username: %v", err)
	}

	return count > 0, nil
}

// InsertUser inserts a new user into the database
func InsertUser(username, difficulty string) (int64, error) {
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

// UpdateUserProgress updates the user's progress (rule reached and time spent)
func UpdateUserProgress(userID int64, ruleReached, timeSpent int) error {
	query := `
		UPDATE users 
		SET rule_reached = ?, time_spent = ?, updated_at = CURRENT_TIMESTAMP
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

// GetUser retrieves a user by ID
func GetUser(userID int64) (*User, error) {
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
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, difficulty, rule_reached, time_spent, created_at, updated_at
		FROM users WHERE username = ?
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
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// GetLeaderboard retrieves the top users ordered by rules reached and time spent
func GetLeaderboard(limit int) ([]User, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT id, username, difficulty, rule_reached, time_spent, created_at, updated_at
		FROM users 
		ORDER BY rule_reached DESC, time_spent ASC
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %v", err)
	}
	defer rows.Close()

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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return users, nil
}

// GetUserStats returns statistics about users
func GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total users
	var totalUsers int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %v", err)
	}
	stats["total_users"] = totalUsers

	// Users by difficulty
	diffQuery := `
		SELECT difficulty, COUNT(*) as count 
		FROM users 
		GROUP BY difficulty 
		ORDER BY count DESC
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
	stats["by_difficulty"] = diffStats

	// Highest rule reached
	var maxRule int
	err = db.QueryRow("SELECT COALESCE(MAX(rule_reached), 0) FROM users").Scan(&maxRule)
	if err != nil {
		return nil, fmt.Errorf("failed to get max rule: %v", err)
	}
	stats["highest_rule"] = maxRule

	// Average time spent
	var avgTime float64
	err = db.QueryRow("SELECT COALESCE(AVG(time_spent), 0) FROM users WHERE time_spent > 0").Scan(&avgTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get average time: %v", err)
	}
	stats["average_time"] = avgTime

	return stats, nil
}

// DeleteUser deletes a user from the database (for testing/admin purposes)
func DeleteUser(userID int64) error {
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
