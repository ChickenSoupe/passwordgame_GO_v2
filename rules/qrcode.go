package rules

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	database "passgame/Database"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

var (
	currentQRWord     string
	currentQRImageB64 string
	qrMutex           sync.RWMutex
)

// QRWord represents a word that can be encoded in a QR code
type QRWord struct {
	ID   int64
	Word string
}

// InitQRCodeTable initializes the QR code words table in the database
func InitQRCodeTable() error {
	db := database.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not available")
	}

	// Create the qr_words table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS qr_words (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT UNIQUE NOT NULL
	);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create qr_words table: %v", err)
	}

	// Check if we need to populate the table with initial words
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM qr_words").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check qr_words count: %v", err)
	}

	// If the table is empty, populate it with some default words
	if count == 0 {
		defaultWords := []string{
			"password", "security", "encryption", "authentication", "verification",
			"protection", "firewall", "cybersecurity", "privacy", "confidential",
			"secret", "hidden", "secure", "private", "locked",
			"key", "code", "cipher", "cryptic", "enigma",
		}

		insertSQL := "INSERT INTO qr_words (word) VALUES (?)"
		for _, word := range defaultWords {
			_, err := db.Exec(insertSQL, word)
			if err != nil {
				log.Printf("Warning: failed to insert QR word '%s': %v", word, err)
				// Continue with other words even if one fails
			}
		}
		log.Println("✅ QR code words table populated with default words")
	}

	return nil
}

// GetRandomQRWord retrieves a random word from the qr_words table
func GetRandomQRWord() (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", fmt.Errorf("database connection not available")
	}

	var word string
	err := db.QueryRow("SELECT word FROM qr_words ORDER BY RANDOM() LIMIT 1").Scan(&word)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no QR words found in database")
		}
		return "", fmt.Errorf("failed to get random QR word: %v", err)
	}

	return word, nil
}

// GenerateQRCode creates a QR code for the given text and returns it as a base64-encoded PNG
func GenerateQRCode(text string) (string, error) {
	// Create the QR code
	qrCode, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return "", fmt.Errorf("failed to encode QR code: %v", err)
	}

	// Scale the QR code to a reasonable size
	qrCode, err = barcode.Scale(qrCode, 200, 200)
	if err != nil {
		return "", fmt.Errorf("failed to scale QR code: %v", err)
	}

	// Convert to PNG
	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, qrCode)
	if err != nil {
		return "", fmt.Errorf("failed to encode QR code as PNG: %v", err)
	}

	// Convert to base64
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// GenerateNewQRCode creates a new QR code with a random word from the database
func GenerateNewQRCode() (string, string, error) {
	word, err := GetRandomQRWord()
	if err != nil {
		return "", "", fmt.Errorf("failed to get random QR word: %v", err)
	}

	qrImageB64, err := GenerateQRCode(word)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	return word, qrImageB64, nil
}

// RefreshQRCode generates a new QR code and updates the current one
func RefreshQRCode() error {
	word, qrImageB64, err := GenerateNewQRCode()
	if err != nil {
		return err
	}

	qrMutex.Lock()
	defer qrMutex.Unlock()

	currentQRWord = word
	currentQRImageB64 = qrImageB64

	return nil
}

// GetCurrentQRWord returns the current QR code word
func GetCurrentQRWord() string {
	qrMutex.RLock()
	defer qrMutex.RUnlock()
	return currentQRWord
}

// GetCurrentQRImageB64 returns the current QR code image as base64
func GetCurrentQRImageB64() string {
	qrMutex.RLock()
	defer qrMutex.RUnlock()
	return currentQRImageB64
}

// ServeQRCodeImage serves the current QR code image
func ServeQRCodeImage(w http.ResponseWriter, r *http.Request) {
	qrMutex.RLock()
	qrImageB64 := currentQRImageB64
	qrMutex.RUnlock()

	if qrImageB64 == "" {
		// Generate new QR code with random string if none exists
		err := RefreshQRCodeWithRandom()
		if err != nil {
			// Fall back to regular refresh if random fails
			err = RefreshQRCode()
			if err != nil {
				http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
				return
			}
		}
		qrMutex.RLock()
		qrImageB64 = currentQRImageB64
		qrMutex.RUnlock()
	}

	// Decode base64 to binary
	imgData, err := base64.StdEncoding.DecodeString(qrImageB64)
	if err != nil {
		http.Error(w, "Invalid QR code image", http.StatusInternalServerError)
		return
	}

	// Prevent caching to ensure fresh images
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.Write(imgData)
}

// RefreshQRCodeHandler generates a new QR code and returns success status
func RefreshQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	// Use the random string generator for refreshing
	err := RefreshQRCodeWithRandom()
	if err != nil {
		// Fall back to regular refresh if random fails
		err = RefreshQRCode()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to refresh QR code: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Get the current word to display in the response
	word := GetCurrentQRWord()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"status": "refreshed", "word": "%s"}`, word)))
}

// ValidateQRCodeWord checks if the password contains the current QR code word
func ValidateQRCodeWord(password string) bool {
	qrMutex.RLock()
	word := currentQRWord
	qrMutex.RUnlock()

	if word == "" {
		return false
	}

	return strings.Contains(strings.ToLower(password), strings.ToLower(word))
}

// GenerateRandomString creates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// AddRandomQRWord adds a new random word to the database
func AddRandomQRWord() (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", fmt.Errorf("database connection not available")
	}

	// Generate a random string between 5-8 characters
	randomWord := GenerateRandomString(rand.Intn(4) + 5)

	// Insert the random word into the database
	insertSQL := "INSERT INTO qr_words (word) VALUES (?) ON CONFLICT(word) DO NOTHING"
	_, err := db.Exec(insertSQL, randomWord)
	if err != nil {
		return "", fmt.Errorf("failed to insert random QR word: %v", err)
	}

	return randomWord, nil
}

// RefreshQRCodeWithRandom generates a new QR code with a random string
func RefreshQRCodeWithRandom() error {
	// Add a new random word to the database
	randomWord, err := AddRandomQRWord()
	if err != nil {
		// If adding a random word fails, fall back to existing words
		return RefreshQRCode()
	}

	// Generate QR code for the random word
	qrImageB64, err := GenerateQRCode(randomWord)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %v", err)
	}

	qrMutex.Lock()
	defer qrMutex.Unlock()

	currentQRWord = randomWord
	currentQRImageB64 = qrImageB64

	return nil
}

// Initialize QR code on package load
func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Initial QR code will be generated when the database is initialized
	// This happens in the main.go file after the database is connected

	// We'll also set up a goroutine to periodically refresh the QR code
	// This ensures users always get a fresh QR code when they reach this rule
	go func() {
		// Wait for database initialization (5 seconds should be enough)
		time.Sleep(5 * time.Second)

		// Refresh the QR code every 10 minutes
		for {
			// Try to refresh with a random string first
			err := RefreshQRCodeWithRandom()
			if err != nil {
				// Fall back to regular refresh if random fails
				_ = RefreshQRCode()
			}

			// Wait before refreshing again
			time.Sleep(10 * time.Minute)
		}
	}()
}
