package main

import (
	"log"
	"net/http"
	"os"

	database "passgame/Database"
	"passgame/component"
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

	// Serve static files from Frontend directory
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "Frontend/style.css")
	})

	http.HandleFunc("/flip-animations.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "Frontend/flip-animations.js")
	})

	log.Println("🚀 Password Game server starting on :8080")
	log.Println("🌐 Open http://localhost:8080 in your browser")
	log.Println("🎮 Password Game: http://localhost:8080/display")
	log.Println("🏆 Leaderboard: http://localhost:8080/leaderboard")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
