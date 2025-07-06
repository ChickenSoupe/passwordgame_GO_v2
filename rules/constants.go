package rules

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	database "passgame/Database"
)

// Global variables to store current mathematical constant and color
var (
	currentConstant     string
	currentConstantName string
	currentColor        string
	currentColorName    string
	constantsMutex      sync.RWMutex
	colorsMutex         sync.RWMutex
)

// MathConstant represents a mathematical constant in the database
type MathConstant struct {
	ID        int64
	Name      string
	Value     string
	ShortDesc string
}

// ColorCode represents a color code in the database
type ColorCode struct {
	ID        int64
	Name      string
	HexCode   string
	ShortDesc string
}

// InitConstantsTable initializes the mathematical constants table in the database
func InitConstantsTable() error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Create the math_constants table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS math_constants (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL,
		short_desc TEXT
	);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create math_constants table: %v", err)
	}

	// Check if we need to populate the table with initial constants
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM math_constants").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check math_constants count: %v", err)
	}

	// If the table is empty, populate it with default constants
	if count == 0 {
		defaultConstants := []MathConstant{
			{Name: "Pi (π)", Value: "3.14159265358979323846", ShortDesc: "Ratio of a circle's circumference to its diameter"},
			{Name: "Euler's Number (e)", Value: "2.71828182845904523536", ShortDesc: "Base of the natural logarithm"},
			{Name: "Golden Ratio (φ)", Value: "1.61803398874989484820", ShortDesc: "Special number where (a+b)/a = a/b"},
			{Name: "Square Root of 2", Value: "1.41421356237309504880", ShortDesc: "Diagonal of a unit square"},
			{Name: "Square Root of 3", Value: "1.73205080756887729352", ShortDesc: "Diagonal of a unit cube"},
			{Name: "Euler-Mascheroni Constant (γ)", Value: "0.57721566490153286060", ShortDesc: "Limit of harmonic series minus natural logarithm"},
			{Name: "Feigenbaum Constant (δ)", Value: "4.66920160910299067185", ShortDesc: "Ratio of successive bifurcation intervals"},
			{Name: "Apéry's Constant (ζ(3))", Value: "1.20205690315959428539", ShortDesc: "Sum of reciprocals of cubes of positive integers"},
			{Name: "Conway's Constant (λ)", Value: "1.30357726903429639125", ShortDesc: "Growth rate of the Look-and-Say sequence"},
			{Name: "Khinchin's Constant (K)", Value: "2.68545200106530644530", ShortDesc: "Geometric mean of continued fraction terms"},
		}

		insertSQL := "INSERT INTO math_constants (name, value, short_desc) VALUES (?, ?, ?)"
		for _, constant := range defaultConstants {
			_, err := db.Exec(insertSQL, constant.Name, constant.Value, constant.ShortDesc)
			if err != nil {
				log.Printf("Warning: failed to insert math constant '%s': %v", constant.Name, err)
				// Continue with other constants even if one fails
			}
		}
		log.Println("✅ Mathematical constants table populated with default values")
	}

	return nil
}

// InitColorsTable initializes the color codes table in the database
func InitColorsTable() error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Create the color_codes table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS color_codes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		hex_code TEXT NOT NULL,
		short_desc TEXT
	);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create color_codes table: %v", err)
	}

	// Check if we need to populate the table with initial colors
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM color_codes").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check color_codes count: %v", err)
	}

	// If the table is empty, populate it with default colors
	if count == 0 {
		defaultColors := []ColorCode{
			{Name: "Red", HexCode: "#FF0000", ShortDesc: "Primary color"},
			{Name: "Green", HexCode: "#00FF00", ShortDesc: "Primary color"},
			{Name: "Blue", HexCode: "#0000FF", ShortDesc: "Primary color"},
			{Name: "Yellow", HexCode: "#FFFF00", ShortDesc: "Secondary color"},
			{Name: "Cyan", HexCode: "#00FFFF", ShortDesc: "Secondary color"},
			{Name: "Magenta", HexCode: "#FF00FF", ShortDesc: "Secondary color"},
			{Name: "Black", HexCode: "#000000", ShortDesc: "Absence of color"},
			{Name: "White", HexCode: "#FFFFFF", ShortDesc: "All colors combined"},
			{Name: "Orange", HexCode: "#FFA500", ShortDesc: "Secondary color"},
			{Name: "Purple", HexCode: "#800080", ShortDesc: "Secondary color"},
			{Name: "Pink", HexCode: "#FFC0CB", ShortDesc: "Light red"},
			{Name: "Brown", HexCode: "#A52A2A", ShortDesc: "Dark orange-red"},
			{Name: "Gray", HexCode: "#808080", ShortDesc: "Neutral color"},
			{Name: "Turquoise", HexCode: "#40E0D0", ShortDesc: "Blue-green color"},
			{Name: "Gold", HexCode: "#FFD700", ShortDesc: "Yellow-orange color"},
			{Name: "Silver", HexCode: "#C0C0C0", ShortDesc: "Metallic gray"},
			{Name: "Navy", HexCode: "#000080", ShortDesc: "Dark blue"},
			{Name: "Teal", HexCode: "#008080", ShortDesc: "Blue-green color"},
			{Name: "Olive", HexCode: "#808000", ShortDesc: "Dark yellow-green"},
			{Name: "Maroon", HexCode: "#800000", ShortDesc: "Dark red"},
		}

		insertSQL := "INSERT INTO color_codes (name, hex_code, short_desc) VALUES (?, ?, ?)"
		for _, color := range defaultColors {
			_, err := db.Exec(insertSQL, color.Name, color.HexCode, color.ShortDesc)
			if err != nil {
				log.Printf("Warning: failed to insert color '%s': %v", color.Name, err)
				// Continue with other colors even if one fails
			}
		}
		log.Println("✅ Color codes table populated with default values")
	}

	return nil
}

// GetRandomMathConstant retrieves a random mathematical constant from the database
func GetRandomMathConstant() (string, string, error) {
	db := database.GetDB()
	if db == nil {
		return "", "", fmt.Errorf("database connection not available")
	}

	var name, value string
	err := db.QueryRow("SELECT name, value FROM math_constants ORDER BY RANDOM() LIMIT 1").Scan(&name, &value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("no mathematical constants found in database")
		}
		return "", "", fmt.Errorf("failed to get random mathematical constant: %v", err)
	}

	return name, value, nil
}

// GetRandomColor retrieves a random color from the database
func GetRandomColor() (string, string, error) {
	db := database.GetDB()
	if db == nil {
		return "", "", fmt.Errorf("database connection not available")
	}

	var name, hexCode string
	err := db.QueryRow("SELECT name, hex_code FROM color_codes ORDER BY RANDOM() LIMIT 1").Scan(&name, &hexCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("no colors found in database")
		}
		return "", "", fmt.Errorf("failed to get random color: %v", err)
	}

	return name, hexCode, nil
}

// RefreshMathConstant generates a new random mathematical constant
func RefreshMathConstant() error {
	name, value, err := GetRandomMathConstant()
	if err != nil {
		return err
	}

	constantsMutex.Lock()
	defer constantsMutex.Unlock()

	currentConstantName = name
	currentConstant = value

	return nil
}

// RefreshColor generates a new random color
func RefreshColor() error {
	name, hexCode, err := GetRandomColor()
	if err != nil {
		return err
	}

	colorsMutex.Lock()
	defer colorsMutex.Unlock()

	currentColorName = name
	currentColor = hexCode

	return nil
}

// GetCurrentMathConstant returns the current mathematical constant
func GetCurrentMathConstant() (string, string) {
	constantsMutex.RLock()
	defer constantsMutex.RUnlock()
	return currentConstantName, currentConstant
}

// GetCurrentColor returns the current color
func GetCurrentColor() (string, string) {
	colorsMutex.RLock()
	defer colorsMutex.RUnlock()
	return currentColorName, currentColor
}

// ValidateMathConstant checks if the password contains the first 3 digits of the current mathematical constant
func ValidateMathConstant(password string) bool {
	constantsMutex.RLock()
	constant := currentConstant
	constantsMutex.RUnlock()

	if constant == "" {
		return false
	}

	// Extract the first 3 digits (ignoring decimal point)
	firstThreeDigits := ""
	digitCount := 0
	for _, char := range constant {
		if char >= '0' && char <= '9' {
			firstThreeDigits += string(char)
			digitCount++
			if digitCount == 3 {
				break
			}
		}
	}

	if len(firstThreeDigits) < 3 {
		return false
	}

	return strings.Contains(password, firstThreeDigits)
}

// ValidateHexColor checks if the password contains the hex code of the current color
func ValidateHexColor(password string) bool {
	colorsMutex.RLock()
	hexCode := currentColor
	colorsMutex.RUnlock()

	if hexCode == "" {
		return false
	}

	// Check for hex code with or without the # prefix
	hexWithoutHash := strings.TrimPrefix(hexCode, "#")

	return strings.Contains(strings.ToLower(password), strings.ToLower(hexCode)) ||
		strings.Contains(strings.ToLower(password), strings.ToLower(hexWithoutHash))
}

// GetMathConstantForHint returns the current mathematical constant for display in hints
func GetMathConstantForHint() string {
	constantsMutex.RLock()
	defer constantsMutex.RUnlock()

	if currentConstantName == "" || currentConstant == "" {
		return "π (3.14159...)"
	}

	// Extract the first 5 digits (including decimal point if present)
	shortValue := currentConstant
	if len(shortValue) > 7 {
		shortValue = shortValue[:7] + "..."
	}

	return fmt.Sprintf("%s (%s)", currentConstantName, shortValue)
}

// GetColorForHint returns the current color for display in hints
func GetColorForHint() string {
	colorsMutex.RLock()
	defer colorsMutex.RUnlock()

	if currentColorName == "" || currentColor == "" {
		return "Red (#FF0000)"
	}

	return fmt.Sprintf("%s (%s)", currentColorName, currentColor)
}

// Initialize constants and colors on package load
func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Initial values will be generated when the database is initialized
	// This happens in the main.go file after the database is connected

	// We'll also set up goroutines to periodically refresh the values
	go func() {
		// Wait for database initialization (5 seconds should be enough)
		time.Sleep(5 * time.Second)

		// Initial refresh
		_ = RefreshMathConstant()
		_ = RefreshColor()

		// Refresh every 6 hours
		for {
			time.Sleep(6 * time.Hour)
			_ = RefreshMathConstant()
			_ = RefreshColor()
		}
	}()
}
