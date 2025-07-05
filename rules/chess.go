package rules

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/corentings/chess/v2"
	chessimage "github.com/corentings/chess/v2/image"
)

// Global variables to store current chess state
var (
	currentChessGame *chess.Game
	currentBestMove  string
	chessMutex       sync.RWMutex
)

// Chess positions for puzzles (FEN notation)
var chessPuzzles = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",              // Starting position
	"r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4",   // Italian Game
	"rnbqkb1r/ppp2ppp/4pn2/3p4/2PP4/2N2N2/PP2PPPP/R1BQKB1R b KQkq - 3 4",    // Queen's Gambit Declined
	"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 4 5", // Spanish Opening
	"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2",         // Scandinavian Defense
}

// getBestMoveFromStockfish gets the best move from Stockfish API
func getBestMoveFromStockfish(fen string) (string, error) {
	// Encode FEN for URL
	encodedFEN := strings.ReplaceAll(fen, " ", "%20")
	url := fmt.Sprintf("https://stockfish.online/api/s/v2.php?fen=%s&depth=15", encodedFEN)
	
	// Set timeout to prevent hanging
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Make API request to Stockfish
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to call Stockfish API: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Stockfish API returned status: %s", resp.Status)
	}

	// Read response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse Stockfish response: %v", err)
	}

	// Extract best move from response (format: "bestmove b7b6 ponder a1e1")
	bestMove, ok := result["bestmove"].(string)
	if !ok || bestMove == "" {
		return "", fmt.Errorf("invalid response from Stockfish")
	}

	// Extract just the move part (e.g., "b7b6" from "bestmove b7b6 ponder a1e1")
	parts := strings.Fields(bestMove)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid move format from Stockfish")
	}
	bestMove = parts[1] // Get the move part (second element)

	// Ensure the move is in the correct format (e.g., e2e4)

	return bestMove, nil
}

// GenerateNewChessPosition creates a new chess position and calculates the best move
func GenerateNewChessPosition() (string, error) {
	chessMutex.Lock()
	defer chessMutex.Unlock()

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Select a random puzzle
	puzzleIndex := rand.Intn(len(chessPuzzles))
	selectedFEN := chessPuzzles[puzzleIndex]

	// Create new game from FEN
	fen, err := chess.FEN(selectedFEN)
	if err != nil {
		return "", fmt.Errorf("failed to parse FEN: %v", err)
	}

	game := chess.NewGame(fen)
	currentChessGame = game

	// Get the best move from Stockfish
	bestMove, err := getBestMoveFromStockfish(selectedFEN)
	if err != nil {
		log.Printf("Failed to get best move from Stockfish: %v, falling back to random move", err)
		// Fallback to random move if Stockfish fails
		moves := game.ValidMoves()
		if len(moves) == 0 {
			return "", fmt.Errorf("no valid moves available")
		}
		bestMove = moves[0].String()
	}

	currentBestMove = bestMove
	return currentBestMove, nil
}

// GetCurrentChessPosition returns the current chess position and best move
func GetCurrentChessPosition() (*chess.Game, string) {
	chessMutex.RLock()
	defer chessMutex.RUnlock()
	return currentChessGame, currentBestMove
}

// generateChessboardImage creates a visual representation of the chess board using the chess/image package
func generateChessboardImage(game *chess.Game) ([]byte, error) {
	// Create a buffer to hold the SVG data
	var buf bytes.Buffer

	// Generate SVG using the chess/image package
	err := chessimage.SVG(&buf, game.Position().Board())
	if err != nil {
		return nil, fmt.Errorf("failed to generate chess board SVG: %v", err)
	}

	return buf.Bytes(), nil
}

// ServeChessImage serves the chess board image
func ServeChessImage(w http.ResponseWriter, r *http.Request) {
	chessMutex.RLock()
	game := currentChessGame
	chessMutex.RUnlock()

	if game == nil {
		// Generate new position if none exists
		_, err := GenerateNewChessPosition()
		if err != nil {
			http.Error(w, "Failed to generate chess position", http.StatusInternalServerError)
			return
		}
		chessMutex.RLock()
		game = currentChessGame
		chessMutex.RUnlock()
	}

	// Generate board SVG using the chess/image package
	svgData, err := generateChessboardImage(game)
	if err != nil {
		http.Error(w, "Failed to generate chess board image", http.StatusInternalServerError)
		return
	}

	// Prevent caching to ensure fresh images
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Serve the SVG data
	w.Write(svgData)
}

// RefreshChess generates a new chess position
func RefreshChess(w http.ResponseWriter, r *http.Request) {
	bestMove, err := GenerateNewChessPosition()
	if err != nil {
		http.Error(w, "Failed to generate new chess position", http.StatusInternalServerError)
		return
	}

	// Return the best move in the response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":   "refreshed",
		"bestMove": bestMove,
	}
	json.NewEncoder(w).Encode(response)
}

// ValidateChessMove checks if the password contains the current best chess move
func ValidateChessMove(password string) bool {
	chessMutex.RLock()
	bestMove := currentBestMove
	chessMutex.RUnlock()

	if bestMove == "" {
		return false
	}

	// Convert password to lowercase for case-insensitive matching
	lowerPassword := strings.ToLower(password)
	lowerBestMove := strings.ToLower(bestMove)

	// Check if the best move is contained in the password
	return strings.Contains(lowerPassword, lowerBestMove)
}

// GetChessBoardAsBase64 returns the current chess board as a base64 encoded SVG
func GetChessBoardAsBase64() (string, error) {
	chessMutex.RLock()
	game := currentChessGame
	chessMutex.RUnlock()

	if game == nil {
		_, err := GenerateNewChessPosition()
		if err != nil {
			return "", err
		}
		chessMutex.RLock()
		game = currentChessGame
		chessMutex.RUnlock()
	}

	// Generate board SVG
	svgData, err := generateChessboardImage(game)
	if err != nil {
		return "", err
	}

	// Encode to base64
	base64Str := base64.StdEncoding.EncodeToString(svgData)
	return "data:image/svg+xml;base64," + base64Str, nil
}

// Initialize chess position on package load
func init() {
	_, err := GenerateNewChessPosition()
	if err != nil {
		log.Printf("Warning: Failed to initialize chess position: %v", err)
	}
}
