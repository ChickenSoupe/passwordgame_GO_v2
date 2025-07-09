package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	database "passgame/Database"
	"passgame/component"
	"passgame/rules"
)

func main() {
	// Initialize database
	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Initialize QR code table
	err = rules.InitQRCodeTable()
	if err != nil {
		log.Fatalf("Failed to initialize QR code table: %v", err)
	}

	// Initialize mathematical constants table
	err = rules.InitConstantsTable()
	if err != nil {
		log.Fatalf("Failed to initialize mathematical constants table: %v", err)
	}

	// Initialize color codes table
	err = rules.InitColorsTable()
	if err != nil {
		log.Fatalf("Failed to initialize color codes table: %v", err)
	}

	// Generate initial QR code with a word from the API
	err = rules.RefreshQRCodeWithAPI()
	if err != nil {
		log.Printf("Warning: Failed to generate initial QR code with API word: %v", err)
		// Fall back to regular refresh if API fails
		err = rules.RefreshQRCode()
		if err != nil {
			log.Printf("Warning: Failed to generate initial QR code: %v", err)
		}
	}

	// Generate initial mathematical constant
	err = rules.RefreshMathConstant()
	if err != nil {
		log.Printf("Warning: Failed to generate initial mathematical constant: %v", err)
	}

	// Generate initial color
	err = rules.RefreshColor()
	if err != nil {
		log.Printf("Warning: Failed to generate initial color: %v", err)
	}

	// Create Database directory if it doesn't exist
	if err := os.MkdirAll("Database", 0755); err != nil {
		log.Printf("Warning: Could not create Database directory: %v", err)
	}

	// Main routes - both root and /display point to the same handler
	http.HandleFunc("/", component.HandlePasswordGame)
	http.HandleFunc("/display", component.HandlePasswordGame)
	http.HandleFunc("/validate", component.HandleValidate)
	http.HandleFunc("/register-user", component.HandleRegisterUser)
	http.HandleFunc("/user-modal.html", component.HandleUserModal) // Now uses template execution
	http.HandleFunc("/leaderboard", component.HandleLeaderboard)

	// Captcha routes
	http.HandleFunc("/captcha.png", rules.ServeCaptchaImage)
	http.HandleFunc("/refresh-captcha", rules.RefreshCaptcha)

	// Chess routes
	http.HandleFunc("/chess.png", rules.ServeChessImage)
	http.HandleFunc("/refresh-chess", rules.RefreshChess)

	// QR code routes
	http.HandleFunc("/qrcode.png", rules.ServeQRCodeImage)
	http.HandleFunc("/refresh-qrcode", rules.RefreshQRCodeHandler)

	// Color routes
	http.HandleFunc("/color.png", ServeColorImage)
	http.HandleFunc("/refresh-color", RefreshColorHandler)

	// Math constant routes
	http.HandleFunc("/refresh-constant", RefreshConstantHandler)

	// Serve static files from Frontend directory
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "Frontend/style.css")
	})

	http.HandleFunc("/flip-animations.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "Frontend/flip-animations.js")
	})

	// Admin API endpoints
	http.HandleFunc("/api/rules/pool", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rules.Pool())
	})

	http.HandleFunc("/api/rules/assignments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			data, err := ioutil.ReadFile("rules/assignments.json")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Could not read assignments"}`))
				return
			}
			w.Write(data)
			return
		case http.MethodPost:
			var assignments map[string][]int
			if err := json.NewDecoder(r.Body).Decode(&assignments); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"Invalid JSON"}`))
				return
			}
			data, err := json.MarshalIndent(assignments, "", "  ")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Could not marshal assignments"}`))
				return
			}
			if err := ioutil.WriteFile("rules/assignments.json", data, 0644); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Could not write assignments"}`))
				return
			}
			w.Write([]byte(`{"status":"ok"}`))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/difficulties", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		difficulties, err := component.LoadDifficulties()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Could not load difficulties"}`))
			return
		}
		json.NewEncoder(w).Encode(difficulties)
	})

	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "Frontend/admin.html")
	})

	// User delete endpoint for Rule 22
	http.HandleFunc("/api/user/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		cookie, err := r.Cookie("user_session")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		session, exists := component.UserSessions[cookie.Value]
		if !exists {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		err = database.DeleteUser(session.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		delete(component.UserSessions, cookie.Value)
		w.WriteHeader(http.StatusOK)
	})

	// Cybersecurity rules routes
	http.HandleFunc("/api/cysec/status", HandleCyberSecurityStatus)
	http.HandleFunc("/api/cysec/update-alert", HandleUpdateAlert)
	http.HandleFunc("/api/cysec/ad-watched", HandleAdWatched)
	http.HandleFunc("/api/cysec/generate-black-squares", HandleGenerateBlackSquares)
	http.HandleFunc("/api/cysec/reset", HandleResetCyberSecurity)

	log.Println("🚀 Password Game server starting on :8080")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Println("🎮 Password Game: http://localhost:8080/display")
	log.Println("🏆 Leaderboard: http://localhost:8080/leaderboard")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// hexToRGB converts a hex color string to RGB values
func hexToRGB(hexColor string) (r, g, b uint8, err error) {
	// Remove the # prefix if present
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Parse the hex color
	if len(hexColor) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color format: %s", hexColor)
	}

	// Parse the RGB values
	rgb, err := strconv.ParseUint(hexColor, 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hexColor)
	}

	// Extract the RGB components
	r = uint8((rgb >> 16) & 0xFF)
	g = uint8((rgb >> 8) & 0xFF)
	b = uint8(rgb & 0xFF)

	return r, g, b, nil
}

// ServeColorImage serves an image of the current color
func ServeColorImage(w http.ResponseWriter, r *http.Request) {
	// Get the current color
	_, hexCode := rules.GetCurrentColor()

	if hexCode == "" {
		// Generate a new color if none exists
		err := rules.RefreshColor()
		if err != nil {
			http.Error(w, "Failed to generate color", http.StatusInternalServerError)
			return
		}
		_, hexCode = rules.GetCurrentColor()
	}

	// Convert hex to RGB
	red, green, blue, err := hexToRGB(hexCode)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid color format: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a new image
	width, height := 200, 200
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill the image with the color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{red, green, blue, 255})
		}
	}

	// Prevent caching to ensure fresh images
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Encode and serve the image
	png.Encode(w, img)
}

// RefreshColorHandler generates a new random color
func RefreshColorHandler(w http.ResponseWriter, r *http.Request) {
	err := rules.RefreshColor()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to refresh color: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the current color for the response
	colorName, hexCode := rules.GetCurrentColor()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":  "refreshed",
		"name":    colorName,
		"hexCode": hexCode,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleCyberSecurityStatus returns the current status of all cybersecurity rules
func HandleCyberSecurityStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status := rules.GetCyberSecurityStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleUpdateAlert handles the update alert for Rule 14
func HandleUpdateAlert(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Mark update alert as shown
		rules.SetUpdateAlertShown(true)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"status":        "shown",
			"update_string": rules.GetUpdateString(),
		}
		json.NewEncoder(w).Encode(response)
	case http.MethodGet:
		// Get update alert status
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"shown":         rules.IsUpdateAlertShown(),
			"update_string": rules.GetUpdateString(),
		}
		json.NewEncoder(w).Encode(response)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HandleAdWatched handles the ad watched status for Rule 23
func HandleAdWatched(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Mark ad as watched
		rules.SetAdWatched(true)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"status":             "watched",
			"raid_unlock_string": rules.GetRaidUnlockString(),
		}
		json.NewEncoder(w).Encode(response)
	case http.MethodGet:
		// Get ad watched status
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"watched":            rules.IsAdWatched(),
			"raid_unlock_string": rules.GetRaidUnlockString(),
		}
		json.NewEncoder(w).Encode(response)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HandleGenerateBlackSquares generates black squares for Rule 24
func HandleGenerateBlackSquares(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	blackSquares := rules.GenerateBlackSquares()
	count := rules.GetBlackSquareCount()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":      "generated",
		"squares":     blackSquares,
		"count":       count,
		"fatal":       count > 12,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleResetCyberSecurity resets all cybersecurity rule states
func HandleResetCyberSecurity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rules.ResetCyberSecurityRules()
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "reset",
	}
	json.NewEncoder(w).Encode(response)
}

// RefreshConstantHandler generates a new random mathematical constant
func RefreshConstantHandler(w http.ResponseWriter, r *http.Request) {
	err := rules.RefreshMathConstant()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to refresh mathematical constant: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the current constant for the response
	constantName, constantValue := rules.GetCurrentMathConstant()

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status": "refreshed",
		"name":   constantName,
		"value":  constantValue,
	}
	json.NewEncoder(w).Encode(response)
}
