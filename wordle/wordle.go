package wordle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// WordleResponse represents the response from NYT Wordle API
type WordleResponse struct {
	Solution string `json:"solution"`
	Print    struct {
		Date string `json:"date"`
	} `json:"print"`
}

// Cache to store today's answer and avoid repeated API calls
type WordleCache struct {
	Answer string
	Date   string
	mu     sync.RWMutex
}

var cache = &WordleCache{}

// GetTodaysAnswer fetches today's Wordle answer from NYT API
func GetTodaysAnswer() (string, error) {
	today := time.Now().Format("2006-01-02")

	// Check cache first
	cache.mu.RLock()
	if cache.Date == today && cache.Answer != "" {
		answer := cache.Answer
		cache.mu.RUnlock()
		return answer, nil
	}
	cache.mu.RUnlock()

	// Fetch from API
	answer, err := fetchWordleAnswer(today)
	if err != nil {
		// If API fails, try fallback methods
		return getFallbackAnswer(today)
	}

	// Update cache
	cache.mu.Lock()
	cache.Answer = answer
	cache.Date = today
	cache.mu.Unlock()

	return answer, nil
}

// fetchWordleAnswer fetches the answer from NYT API
func fetchWordleAnswer(date string) (string, error) {
	url := fmt.Sprintf("https://www.nytimes.com/svc/wordle/v2/%s.json", date)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to mimic browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://www.nytimes.com/games/wordle/")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch wordle data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var wordleResp WordleResponse
	if err := json.NewDecoder(resp.Body).Decode(&wordleResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if wordleResp.Solution == "" {
		return "", fmt.Errorf("no solution found in response")
	}

	return strings.ToUpper(wordleResp.Solution), nil
}

// getFallbackAnswer provides a fallback method when API fails
func getFallbackAnswer(date string) (string, error) {
	// Calculate Wordle number based on date
	// Wordle #1 was on June 19, 2021
	startDate := time.Date(2021, 6, 19, 0, 0, 0, 0, time.UTC)
	currentDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %w", err)
	}

	daysDiff := int(currentDate.Sub(startDate).Hours() / 24)
	wordleNumber := daysDiff + 1

	// Fallback word list (a subset of common 5-letter words)
	// In a real implementation, you might want a more comprehensive list
	fallbackWords := []string{
		"SLATE", "ROAST", "PRIDE", "STEAM", "HORSE", "DANCE", "LIGHT", "CLOUD", "STONE", "HEART",
		"PLANT", "SWEET", "WORLD", "SMILE", "PEACE", "DREAM", "FLAME", "BRAVE", "SHINE", "GRACE",
		"MOUNT", "BEACH", "FRESH", "CRISP", "HAPPY", "MAGIC", "POWER", "CHARM", "QUIET", "BLOOM",
		"SPARK", "GLEAM", "TREND", "FLASH", "GLORY", "HONEY", "JUICY", "KNEEL", "LUNAR", "MERRY",
		"NOBLE", "OCEAN", "PLUSH", "QUEST", "ROYAL", "SUNNY", "TIGER", "URBAN", "VIVID", "WINDY",
	}

	// Use modulo to cycle through the fallback words
	wordIndex := wordleNumber % len(fallbackWords)
	return fallbackWords[wordIndex], nil
}

// ValidateWordleAnswer checks if the password contains today's Wordle answer
func ValidateWordleAnswer(password string) bool {
	answer, err := GetTodaysAnswer()
	if err != nil {
		// If we can't get the answer, default to a known word for testing
		answer = "SLATE"
	}

	// Check if password contains the wordle answer (case-insensitive)
	return strings.Contains(strings.ToUpper(password), strings.ToUpper(answer))
}

// GetTodaysAnswerForHint returns today's answer for display in hints
func GetTodaysAnswerForHint() string {
	answer, err := GetTodaysAnswer()
	if err != nil {
		return "SLATE" // fallback
	}
	return answer
}
