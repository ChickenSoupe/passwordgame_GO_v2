package fun

import (
	"net/http"
	"sync"

	"github.com/dchest/captcha"
)

// Global variables to store current captcha
var (
	currentCaptchaID string
	captchaMutex     sync.RWMutex
)

// CustomCaptchaStore implements a custom store that doesn't expire captchas
type CustomCaptchaStore struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewCustomCaptchaStore() *CustomCaptchaStore {
	return &CustomCaptchaStore{
		data: make(map[string][]byte),
	}
}

func (s *CustomCaptchaStore) Set(id string, digits []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = make([]byte, len(digits))
	copy(s.data[id], digits)
}

func (s *CustomCaptchaStore) Get(id string, clear bool) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	digits, exists := s.data[id]
	if !exists {
		return nil
	}
	// Don't clear on get - this allows multiple verification attempts
	result := make([]byte, len(digits))
	copy(result, digits)
	return result
}

func (s *CustomCaptchaStore) Collect() {
	// Don't collect anything - keep captchas indefinitely
}

// GenerateNewCaptcha creates a new captcha and returns the ID
func GenerateNewCaptcha() string {
	captchaMutex.Lock()
	defer captchaMutex.Unlock()

	// Create captcha ID with 5 digits
	currentCaptchaID = captcha.NewLen(5)

	return currentCaptchaID
}

// GetCurrentCaptchaID returns the current captcha ID
func GetCurrentCaptchaID() string {
	captchaMutex.RLock()
	defer captchaMutex.RUnlock()
	return currentCaptchaID
}

// ServeCaptchaImage serves the captcha image
func ServeCaptchaImage(w http.ResponseWriter, r *http.Request) {
	captchaMutex.RLock()
	captchaID := currentCaptchaID
	captchaMutex.RUnlock()

	if captchaID == "" {
		// Generate new captcha if none exists
		captchaID = GenerateNewCaptcha()
	}

	// Prevent caching to ensure fresh images
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Always use the current captcha ID to serve the image
	// This ensures the image stays consistent with the validation
	captcha.WriteImage(w, captchaID, captcha.StdWidth, captcha.StdHeight)
}

// RefreshCaptcha generates a new captcha
func RefreshCaptcha(w http.ResponseWriter, r *http.Request) {
	GenerateNewCaptcha()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "refreshed"}`))
}

// ValidateCaptcha checks if the password contains the current captcha solution
func ValidateCaptcha(password string) bool {
	captchaMutex.RLock()
	captchaID := currentCaptchaID
	captchaMutex.RUnlock()

	if captchaID == "" {
		return false
	}

	// Extract all 5-digit sequences from the password and check if any match the captcha
	for i := 0; i <= len(password)-5; i++ {
		candidate := password[i : i+5]
		// Check if this 5-character substring is all digits
		allDigits := true
		for _, char := range candidate {
			if char < '0' || char > '9' {
				allDigits = false
				break
			}
		}
		if allDigits && captcha.VerifyString(captchaID, candidate) {
			return true
		}
	}

	return false
}

// Initialize captcha on package load
func init() {
	// Set custom store that doesn't expire captchas
	captcha.SetCustomStore(NewCustomCaptchaStore())
	GenerateNewCaptcha()
}
