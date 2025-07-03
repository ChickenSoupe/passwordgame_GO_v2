package rules

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/notnil/chess"
)

// Global variables to store current chess state
var (
	currentChessGame *chess.Game
	currentBestMove  string
	chessMutex       sync.RWMutex
)

// Chess positions for puzzles (FEN notation)
var chessPuzzles = []string{
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", // Starting position
	"r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 4 4", // Italian Game
	"rnbqkb1r/ppp2ppp/4pn2/3p4/2PP4/2N2N2/PP2PPPP/R1BQKB1R b KQkq - 3 4", // Queen's Gambit Declined
	"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/3P1N2/PPP2PPP/RNBQK2R w KQkq - 4 5", // Spanish Opening
	"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6 0 2", // Scandinavian Defense
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

	// Get all legal moves
	moves := game.ValidMoves()
	if len(moves) == 0 {
		return "", fmt.Errorf("no valid moves available")
	}

	// For simplicity, we'll use the first legal move as the "best" move
	// In a real implementation, you'd use a chess engine to find the actual best move
	bestMove := moves[0]
	currentBestMove = bestMove.String()

	return currentBestMove, nil
}

// GetCurrentChessPosition returns the current chess position and best move
func GetCurrentChessPosition() (*chess.Game, string) {
	chessMutex.RLock()
	defer chessMutex.RUnlock()
	return currentChessGame, currentBestMove
}

// generateChessboardImage creates a visual representation of the chess board
func generateChessboardImage(game *chess.Game) (image.Image, error) {
	// Create a simple 8x8 board image (400x400 pixels)
	const squareSize = 50
	const boardSize = 8 * squareSize
	
	img := image.NewRGBA(image.Rect(0, 0, boardSize, boardSize))
	
	// Colors for the board
	lightColor := color.RGBA{240, 217, 181, 255} // Light squares
	darkColor := color.RGBA{181, 136, 99, 255}   // Dark squares
	
	// Draw the board squares
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			// Determine square color
			isLight := (row+col)%2 == 0
			squareColor := darkColor
			if isLight {
				squareColor = lightColor
			}
			
			// Fill the square
			for y := row * squareSize; y < (row+1)*squareSize; y++ {
				for x := col * squareSize; x < (col+1)*squareSize; x++ {
					img.Set(x, y, squareColor)
				}
			}
		}
	}
	
	// Get board position
	position := game.Position()
	board := position.Board()
	
	// Draw pieces (simplified representation using colored circles)
	for square := 0; square < 64; square++ {
		piece := board.Piece(chess.Square(square))
		if piece != chess.NoPiece {
			// Calculate position (flip row for proper display)
			row := 7 - (square / 8) // Flip the row
			col := square % 8
			
			// For simplicity, we'll just mark occupied squares with a different color
			// In a real implementation, you'd draw actual piece symbols
			pieceColor := color.RGBA{255, 0, 0, 255} // Red for white pieces
			if piece.Color() == chess.Black {
				pieceColor = color.RGBA{0, 0, 255, 255} // Blue for black pieces
			}
			
			// Draw a small circle to represent the piece
			centerX := col*squareSize + squareSize/2
			centerY := row*squareSize + squareSize/2
			radius := squareSize / 4
			
			for y := centerY - radius; y <= centerY + radius; y++ {
				for x := centerX - radius; x <= centerX + radius; x++ {
					if x >= 0 && x < boardSize && y >= 0 && y < boardSize {
						dx := x - centerX
						dy := y - centerY
						if dx*dx + dy*dy <= radius*radius {
							img.Set(x, y, pieceColor)
						}
					}
				}
			}
		}
	}
	
	return img, nil
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

	// Generate board image
	img, err := generateChessboardImage(game)
	if err != nil {
		http.Error(w, "Failed to generate chess board image", http.StatusInternalServerError)
		return
	}

	// Prevent caching to ensure fresh images
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Encode and serve the image
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes())
}

// RefreshChess generates a new chess position
func RefreshChess(w http.ResponseWriter, r *http.Request) {
	_, err := GenerateNewChessPosition()
	if err != nil {
		http.Error(w, "Failed to generate new chess position", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "refreshed"}`))
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

// GetChessBoardAsBase64 returns the current chess board as a base64 encoded image
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

	// Generate board image
	img, err := generateChessboardImage(game)
	if err != nil {
		return "", err
	}

	// Encode to base64
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + base64Str, nil
}

// Initialize chess position on package load
func init() {
	_, err := GenerateNewChessPosition()
	if err != nil {
		log.Printf("Warning: Failed to initialize chess position: %v", err)
	}
}