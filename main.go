package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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

	log.Println("🚀 Password Game server starting on :8080")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Println("🎮 Password Game: http://localhost:8080/display")
	log.Println("🏆 Leaderboard: http://localhost:8080/leaderboard")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
