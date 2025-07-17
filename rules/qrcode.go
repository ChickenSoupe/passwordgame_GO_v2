package rules

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
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

// FetchRandomWord fetches a random word from multiple APIs with fallback
func FetchRandomWord() (string, error) {
	// Try multiple APIs in order
	apis := []struct {
		name   string
		url    string
		parser func([]byte) (string, error)
	}{
		{
			name: "random-word-api.herokuapp.com",
			url:  "https://random-word-api.herokuapp.com/word",
			parser: func(body []byte) (string, error) {
				var words []string
				if err := json.Unmarshal(body, &words); err != nil {
					return "", fmt.Errorf("failed to parse API response: %v", err)
				}
				if len(words) == 0 {
					return "", fmt.Errorf("API returned empty word list")
				}
				return words[0], nil
			},
		},
		{
			name: "api.wordnik.com",
			url:  "https://api.wordnik.com/v4/words.json/randomWord?hasDictionaryDef=true&minCorpusCount=0&maxCorpusCount=-1&minDictionaryCount=1&maxDictionaryCount=-1&minLength=3&maxLength=15&api_key=a2a73e7b926c924fad7001ca3111acd55af2ffabf50eb4ae5",
			parser: func(body []byte) (string, error) {
				var result struct {
					Word string `json:"word"`
				}
				if err := json.Unmarshal(body, &result); err != nil {
					return "", fmt.Errorf("failed to parse API response: %v", err)
				}
				if result.Word == "" {
					return "", fmt.Errorf("API returned empty word")
				}
				return result.Word, nil
			},
		},
	}

	for _, api := range apis {
		word, err := fetchRandomWordFromAPI(api.url, api.parser)
		if err == nil {
			return word, nil
		}
		log.Printf("API %s failed: %v", api.name, err)
	}

	return "", fmt.Errorf("all APIs failed")
}

// fetchRandomWordFromAPI attempts to fetch a word from a specific API
func fetchRandomWordFromAPI(apiURL string, parser func([]byte) (string, error)) (string, error) {
	return fetchRandomWordWithRetry(apiURL, parser, 2, 2*time.Second)
}

// fetchRandomWordWithRetry attempts to fetch a random word with exponential backoff
func fetchRandomWordWithRetry(apiURL string, parser func([]byte) (string, error), maxRetries int, initialDelay time.Duration) (string, error) {
	// Create a client with a timeout to prevent hanging
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var lastErr error
	delay := initialDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Make the request
		resp, err := client.Get(apiURL)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch random word from API: %v", err)
			if attempt < maxRetries-1 {
				log.Printf("API attempt %d failed, retrying in %v: %v", attempt+1, delay, err)
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
				continue
			}
			return "", lastErr
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("API returned non-OK status: %d", resp.StatusCode)
			if attempt < maxRetries-1 {
				log.Printf("API attempt %d failed with status %d, retrying in %v", attempt+1, resp.StatusCode, delay)
				time.Sleep(delay)
				delay *= 2
				continue
			}
			return "", lastErr
		}

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read API response: %v", err)
			if attempt < maxRetries-1 {
				log.Printf("API attempt %d failed to read response, retrying in %v: %v", attempt+1, delay, err)
				time.Sleep(delay)
				delay *= 2
				continue
			}
			return "", lastErr
		}

		// Parse the JSON response using the provided parser
		word, err := parser(body)
		if err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				log.Printf("API attempt %d failed to parse response, retrying in %v: %v", attempt+1, delay, err)
				time.Sleep(delay)
				delay *= 2
				continue
			}
			return "", lastErr
		}

		// Success! Return the word
		return word, nil
	}

	return "", lastErr
}

// GetFallbackWords returns a list of fallback words in case the API is unavailable
func GetFallbackWords() []string {
	return []string{
		// Security-related words
		"password", "security", "encryption", "authentication", "verification",
		"protection", "firewall", "cybersecurity", "privacy", "confidential",
		"secret", "hidden", "secure", "private", "locked", "key", "code",
		"token", "access", "login", "session", "certificate", "signature",

		// Technology words
		"computer", "keyboard", "mouse", "monitor", "server", "database",
		"network", "internet", "software", "hardware", "system", "program",
		"website", "browser", "application", "platform", "framework", "library",

		// Nature words
		"tiger", "lion", "elephant", "dolphin", "eagle", "penguin", "turtle",
		"mountain", "ocean", "beach", "forest", "jungle", "desert", "island",
		"river", "lake", "tree", "flower", "grass", "cloud", "sunshine",

		// Food words
		"apple", "banana", "orange", "pizza", "coffee", "bread", "cheese",
		"chicken", "salmon", "pasta", "rice", "chocolate", "cookie", "cake",

		// Common adjectives
		"happy", "amazing", "awesome", "fantastic", "brilliant", "beautiful",
		"wonderful", "incredible", "magnificent", "spectacular", "excellent",
		"perfect", "outstanding", "remarkable", "extraordinary", "fabulous",

		// Common nouns
		"house", "car", "book", "phone", "music", "movie", "game", "sport",
		"travel", "journey", "adventure", "dream", "story", "memory", "friend",
		"family", "love", "hope", "peace", "joy", "success", "future",

		// Action words
		"create", "build", "design", "develop", "explore", "discover", "learn",
		"teach", "share", "connect", "communicate", "innovate", "inspire",
		"achieve", "accomplish", "complete", "finish", "start", "begin",
	}
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

	// If the table is empty, populate it with fallback words
	if count == 0 {
		fallbackWords := GetFallbackWords()

		insertSQL := "INSERT INTO qr_words (word) VALUES (?)"
		for _, word := range fallbackWords {
			_, err := db.Exec(insertSQL, word)
			if err != nil {
				log.Printf("Warning: failed to insert QR word '%s': %v", word, err)
				// Continue with other words even if one fails
			}
		}
		log.Println("✅ QR code words table populated with fallback words")
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
		// Generate new QR code with a word from the API if none exists
		err := RefreshQRCodeWithAPI()
		if err != nil {
			// Fall back to regular refresh if API word generation fails
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
	// Use the API word generator for refreshing
	err := RefreshQRCodeWithAPI()
	if err != nil {
		// Fall back to regular refresh if API word generation fails
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

// AddRandomWordFromAPI adds a new random word from the API to the database
func AddRandomWordFromAPI() (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", fmt.Errorf("database connection not available")
	}

	// Fetch a random word from the API
	randomWord, err := FetchRandomWord()
	if err != nil {
		// If API fails, fall back to a random word from our fallback list
		log.Printf("Warning: Failed to fetch word from API: %v. Using fallback.", err)
		fallbackWords := GetFallbackWords()
		randomWord = fallbackWords[rand.Intn(len(fallbackWords))]
	}

	// Insert the word into the database if it doesn't exist
	insertSQL := "INSERT INTO qr_words (word) VALUES (?) ON CONFLICT(word) DO NOTHING"
	_, err = db.Exec(insertSQL, randomWord)
	if err != nil {
		return "", fmt.Errorf("failed to insert random QR word: %v", err)
	}

	return randomWord, nil
}

// RefreshQRCodeWithAPI generates a new QR code with a word from the API
func RefreshQRCodeWithAPI() error {
	// Add a new word from the API to the database
	apiWord, err := AddRandomWordFromAPI()
	if err != nil {
		// If adding an API word fails, fall back to existing words
		return RefreshQRCode()
	}

	// Generate QR code for the API word
	qrImageB64, err := GenerateQRCode(apiWord)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %v", err)
	}

	qrMutex.Lock()
	defer qrMutex.Unlock()

	currentQRWord = apiWord
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
			// Try to refresh with a word from the API first
			err := RefreshQRCodeWithAPI()
			if err != nil {
				// Fall back to regular refresh if API word generation fails
				_ = RefreshQRCode()
			}

			// Wait before refreshing again
			time.Sleep(10 * time.Minute)
		}
	}()
}
