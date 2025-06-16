package component

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	database "passgame/Database"
)

// LeaderboardData holds data for the leaderboard template
type LeaderboardData struct {
	Title    string
	Users    []database.User
	Stats    map[string]interface{}
	HasUsers bool
	ErrorMsg string
}

// HandleLeaderboard handles the leaderboard page
func HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Get leaderboard data
	users, err := database.GetLeaderboard(50) // Top 50 users
	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		handleLeaderboardError(w, "Failed to load leaderboard data")
		return
	}

	// Get user statistics
	stats, err := database.GetUserStats()
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
		stats = make(map[string]interface{}) // Empty stats on error
	}

	data := LeaderboardData{
		Title:    "Password Game - Leaderboard",
		Users:    users,
		Stats:    stats,
		HasUsers: len(users) > 0,
	}

	// Parse and execute template
	tmpl, err := template.New("leaderboard").Funcs(template.FuncMap{
		"formatDuration":     formatDuration,
		"formatTime":         formatTime,
		"getRank":            getRank,
		"getDifficultyIcon":  getDifficultyIcon,
		"getDifficultyColor": getDifficultyColor,
	}).Parse(leaderboardTemplate)

	if err != nil {
		log.Printf("Error parsing leaderboard template: %v", err)
		handleLeaderboardError(w, "Template error")
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing leaderboard template: %v", err)
		handleLeaderboardError(w, "Failed to render page")
		return
	}
}

// handleLeaderboardError handles errors by showing a simple HTML error page
func handleLeaderboardError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusInternalServerError)

	errorHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Error - Password Game Leaderboard</title>
    <link rel="stylesheet" href="/style.css">
</head>
<body>
    <div class="container">
        <div class="error-container">
            <h1>⚠️ Error</h1>
            <p>%s</p>
            <a href="/" class="btn-primary">← Back to Game</a>
        </div>
    </div>
</body>
</html>`, message)

	w.Write([]byte(errorHTML))
}

// Template helper functions
func formatDuration(seconds int) string {
	if seconds == 0 {
		return "0s"
	}

	duration := time.Duration(seconds) * time.Second

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		if seconds == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

func formatTime(t time.Time) string {
	return t.Format("Jan 2, 2006")
}

func getRank(index int) int {
	return index + 1
}

func getDifficultyIcon(difficulty string) string {
	switch difficulty {
	case "basic":
		return "🟢"
	case "intermediate":
		return "🟡"
	case "hard":
		return "🔴"
	case "fun":
		return "🎉"
	default:
		return "⚪"
	}
}

func getDifficultyColor(difficulty string) string {
	switch difficulty {
	case "basic":
		return "#4ade80"
	case "intermediate":
		return "#facc15"
	case "hard":
		return "#f87171"
	case "fun":
		return "#a78bfa"
	default:
		return "#64748b"
	}
}

// leaderboardTemplate is the HTML template for the leaderboard
const leaderboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/style.css">
    <style>
        .leaderboard-container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 2rem;
        }
        
        .leaderboard-title {
            font-size: 3rem;
            text-align: center;
            margin-bottom: 2rem;
            background: linear-gradient(45deg, #00d4ff, #0099cc);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        
        .stat-card {
            background: rgba(255, 255, 255, 0.1);
            padding: 1.5rem;
            border-radius: 12px;
            text-align: center;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }
        
        .stat-value {
            font-size: 2rem;
            font-weight: bold;
            color: #00d4ff;
        }
        
        .stat-label {
            color: rgba(255, 255, 255, 0.8);
            margin-top: 0.5rem;
        }
        
        .leaderboard-table {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            overflow: hidden;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .table-header {
            background: rgba(0, 212, 255, 0.2);
            padding: 1rem;
            display: grid;
            grid-template-columns: 60px 1fr 120px 100px 100px 120px;
            gap: 1rem;
            font-weight: bold;
            color: white;
        }
        
        .table-row {
            padding: 1rem;
            display: grid;
            grid-template-columns: 60px 1fr 120px 100px 100px 120px;
            gap: 1rem;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
            transition: background 0.3s ease;
            align-items: center;
        }
        
        .table-row:hover {
            background: rgba(255, 255, 255, 0.05);
        }
        
        .table-row:last-child {
            border-bottom: none;
        }
        
        .rank {
            font-weight: bold;
            font-size: 1.2rem;
        }
        
        .rank.gold { color: #ffd700; }
        .rank.silver { color: #c0c0c0; }
        .rank.bronze { color: #cd7f32; }
        
        .username {
            font-weight: 600;
            color: white;
        }
        
        .difficulty-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.25rem 0.75rem;
            border-radius: 20px;
            font-size: 0.9rem;
            font-weight: 500;
        }
        
        .rule-progress {
            font-weight: bold;
            color: #00d4ff;
        }
        
        .time-spent {
            color: rgba(255, 255, 255, 0.8);
        }
        
        .join-date {
            color: rgba(255, 255, 255, 0.6);
            font-size: 0.9rem;
        }
        
        .empty-state {
            text-align: center;
            padding: 4rem 2rem;
            color: rgba(255, 255, 255, 0.6);
        }
        
        .empty-state h3 {
            font-size: 1.5rem;
            margin-bottom: 1rem;
            color: rgba(255, 255, 255, 0.8);
        }
        
        @media (max-width: 768px) {
            .table-header,
            .table-row {
                grid-template-columns: 1fr;
                gap: 0.5rem;
            }
            
            .table-header {
                display: none;
            }
            
            .table-row {
                padding: 1.5rem 1rem;
                text-align: center;
            }
            
            .leaderboard-title {
                font-size: 2rem;
            }
        }
    </style>
</head>
<body>
    <!-- Sidebar Toggle -->
    <input type="checkbox" id="navcheck" role="button" title="menu">
    <label for="navcheck" aria-hidden="true" title="menu">
        <span class="burger">
            <span class="bar">
                <span class="visuallyhidden">Menu</span>
            </span>
        </span>
    </label>
    
    <!-- Sidebar Navigation -->
    <nav id="menu">
        <a href="/">
            <span class="menu-icon">🏠</span>
            <span class="menu-text">Password Game</span>
        </a>
        <a href="/leaderboard">
            <span class="menu-icon">🏆</span>
            <span class="menu-text">Leaderboard</span>
        </a>
    </nav>
    
    <main>
        <div class="content">
            <div class="leaderboard-container">
                <h1 class="leaderboard-title">🏆 Leaderboard</h1>
                
                {{if .Stats}}
                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-value">{{.Stats.total_users}}</div>
                        <div class="stat-label">Total Players</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">{{.Stats.highest_rule}}</div>
                        <div class="stat-label">Highest Rule Reached</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-value">{{printf "%.0f" .Stats.average_time}}s</div>
                        <div class="stat-label">Average Time</div>
                    </div>
                </div>
                {{end}}
                
                {{if .HasUsers}}
                <div class="leaderboard-table">
                    <div class="table-header">
                        <div>Rank</div>
                        <div>Player</div>
                        <div>Difficulty</div>
                        <div>Rules</div>
                        <div>Time</div>
                        <div>Joined</div>
                    </div>
                    
                    {{range $index, $user := .Users}}
                    <div class="table-row">
                        <div class="rank {{if eq (getRank $index) 1}}gold{{else if eq (getRank $index) 2}}silver{{else if eq (getRank $index) 3}}bronze{{end}}">
                            #{{getRank $index}}
                        </div>
                        <div class="username">{{$user.Username}}</div>
                        <div>
                            <span class="difficulty-badge" style="background-color: {{getDifficultyColor $user.Difficulty}}20; color: {{getDifficultyColor $user.Difficulty}};">
                                {{getDifficultyIcon $user.Difficulty}} {{$user.Difficulty}}
                            </span>
                        </div>
                        <div class="rule-progress">{{$user.RuleReached}}/20</div>
                        <div class="time-spent">{{formatDuration $user.TimeSpent}}</div>
                        <div class="join-date">{{formatTime $user.CreatedAt}}</div>
                    </div>
                    {{end}}
                </div>
                {{else}}
                <div class="empty-state">
                    <h3>🎮 No Players Yet!</h3>
                    <p>Be the first to join the Password Game challenge!</p>
                    <a href="/" class="btn-primary" style="margin-top: 1rem; display: inline-block;">Start Playing</a>
                </div>
                {{end}}
            </div>
        </div>
    </main>
</body>
</html>`
