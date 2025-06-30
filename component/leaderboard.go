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
	Title     string
	Users     []database.User
	Stats     map[string]interface{}
	HasUsers  bool
	ErrorMsg  string
	SortBy    string
	SortOrder string
	IsHtmx    bool
}

// HandleLeaderboard handles the leaderboard page
func HandleLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Check if this is an HTMX request
	isHtmx := r.Header.Get("HX-Request") == "true"

	// Get sort parameters from URL with defaults
	sortBy := getQueryParam(r, "sort", "rule")
	sortOrder := getQueryParam(r, "order", "desc")
	difficulty := getQueryParam(r, "difficulty", "")

	// Get leaderboard data with sorting and filtering
	var users []database.User
	var err error

	if difficulty != "" && difficulty != "all" {
		users, err = database.GetLeaderboardByDifficulty(difficulty, 20, sortBy, sortOrder)
	} else {
		users, err = database.GetLeaderboardSorted(20, sortBy, sortOrder)
	}

	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		handleLeaderboardError(w, "Failed to load leaderboard data", isHtmx)
		return
	}

	// Prepare data for template
	data := LeaderboardData{
		Title:     "Password Game - Leaderboard",
		Users:     users,
		HasUsers:  len(users) > 0,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		IsHtmx:    isHtmx,
	}

	// For full page loads, get additional stats
	if !isHtmx {
		stats, err := database.GetUserStats()
		if err != nil {
			log.Printf("Error getting user stats: %v", err)
			stats = make(map[string]interface{})
		}
		data.Stats = stats
	}

	// Create template with proper parsing
	if isHtmx {
		// For HTMX requests, return only the table content
		renderLeaderboardTable(w, data)
	} else {
		// For full page requests, render the complete page
		renderFullLeaderboard(w, data)
	}
}

// renderLeaderboardTable renders just the table for HTMX requests
func renderLeaderboardTable(w http.ResponseWriter, data LeaderboardData) {
	tmpl := template.New("leaderboard-table").Funcs(getTemplateFunctions())

	tmpl, err := tmpl.Parse(leaderboardTableTemplate)
	if err != nil {
		log.Printf("Error parsing table template: %v", err)
		handleLeaderboardError(w, "Template error", true)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing table template: %v", err)
		handleLeaderboardError(w, "Failed to render table", true)
	}
}

// renderFullLeaderboard renders the complete page
func renderFullLeaderboard(w http.ResponseWriter, data LeaderboardData) {
	tmpl := template.New("leaderboard").Funcs(getTemplateFunctions())

	// Parse both templates
	tmpl, err := tmpl.Parse(leaderboardTemplate)
	if err != nil {
		log.Printf("Error parsing main template: %v", err)
		handleLeaderboardError(w, "Template error", false)
		return
	}

	// Parse the table template as well
	tmpl, err = tmpl.Parse(leaderboardTableTemplate)
	if err != nil {
		log.Printf("Error parsing table template: %v", err)
		handleLeaderboardError(w, "Template error", false)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing main template: %v", err)
		handleLeaderboardError(w, "Failed to render page", false)
	}
}

// getTemplateFunctions returns all template helper functions
func getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"formatDuration":     formatDuration,
		"formatTime":         formatTime,
		"getRank":            getRank,
		"getDifficultyIcon":  getDifficultyIcon,
		"getDifficultyColor": getDifficultyColor,
		"getSortIcon":        getSortIcon,
		"toggleSortOrder":    toggleSortOrder,
		"getNextDifficulty":  getNextDifficulty,
	}
}

// getQueryParam safely gets a query parameter with a default value
func getQueryParam(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// handleLeaderboardError handles errors appropriately for both full and partial requests
func handleLeaderboardError(w http.ResponseWriter, message string, isHtmx bool) {
	if isHtmx {
		w.Header().Set("HX-Reswap", "none")
		w.Header().Set("HX-Retarget", "#error-message")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `<div id="error-message" class="error-message">%s</div>`, message)
		return
	}

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

func getSortIcon(currentSort, columnSort, currentOrder string) string {
	if currentSort != columnSort {
		return "↕️"
	}
	if currentOrder == "desc" {
		return "↓"
	}
	return "↑"
}

func toggleSortOrder(currentSort, columnSort, currentOrder string) string {
	if currentSort != columnSort {
		return "desc"
	}
	if currentOrder == "desc" {
		return "asc"
	}
	return "desc"
}

// getNextDifficulty cycles through difficulty filters
func getNextDifficulty(currentDifficulty string) string {
	switch currentDifficulty {
	case "all":
		return "basic"
	case "basic":
		return "intermediate"
	case "intermediate":
		return "hard"
	case "hard":
		return "fun"
	case "fun":
		return "all"
	default:
		return "basic"
	}
}

// leaderboardTemplate is the HTML template for the full leaderboard page
const leaderboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <link rel="stylesheet" href="/style.css">
    <style>
        .sortable-header {
            cursor: pointer;
            user-select: none;
            transition: background-color 0.2s ease;
            position: relative;
            padding: 8px 12px;
        }
        
        .sortable-header:hover {
            background-color: rgba(255, 255, 255, 0.1);
        }
        
        .sorting-active {
            background-color: rgba(255, 255, 255, 0.2) !important;
        }
        
        .sort-indicator {
            position: absolute;
            right: 5px;
            top: 50%;
            transform: translateY(-50%);
            font-size: 12px;
            opacity: 0.7;
        }
        
        .htmx-request .sort-indicator {
            animation: spin 1s linear infinite;
        }
        
        @keyframes spin {
            from { transform: translateY(-50%) rotate(0deg); }
            to { transform: translateY(-50%) rotate(360deg); }
        }
        
        .difficulty-filter {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 4px;
            padding: 4px 8px;
            font-size: 12px;
            margin-left: 8px;
        }

        .active-sort {
            background-color: rgba(255, 255, 255, 0.15);
        }

        .table-responsive {
            overflow-x: auto;
        }

        .error-message {
            background: #fee;
            color: #c33;
            padding: 12px;
            border-radius: 4px;
            margin: 16px 0;
            text-align: center;
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
                <h1 class="leaderboard-title">🏆 Leaderboard (Top 20)</h1>
                
                {{if .Stats}}
                <!-- Stats Overview -->
                <div class="stats-overview">
                    <div class="stat-item">
                        <div class="stat-value">{{.Stats.total_users}}</div>
                        <div class="stat-label">Total Players</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">{{.Stats.highest_rule}}</div>
                        <div class="stat-label">Highest Rule Reached</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">{{printf "%.0f" .Stats.average_time}}s</div>
                        <div class="stat-label">Average Time</div>
                    </div>
                </div>
                
                <!-- Charts Section -->
                <div class="charts-container">
                    <div class="chart-card">
                        <h3 class="chart-title">📊 Players by Difficulty</h3>
                        <div class="chart-container">
                            <canvas id="difficultyChart"></canvas>
                        </div>
                    </div>
                    
                    <div class="chart-card">
                        <h3 class="chart-title">📈 Rule Progress Distribution</h3>
                        <div class="chart-container">
                            <canvas id="progressChart"></canvas>
                        </div>
                    </div>
                </div>
                {{end}}
                
                <!-- Error message container -->
                <div id="error-message"></div>
                
                <!-- Leaderboard Content -->
                <div id="leaderboard-content" class="table-responsive">
                    {{template "leaderboard-table" .}}
                </div>
            </div>
        </div>
    </main>

    <script>
        // Store current state
        let currentSort = '{{.SortBy}}';
        let currentOrder = '{{.SortOrder}}';
        let currentDifficulty = 'all';
        
        document.addEventListener('DOMContentLoaded', function() {
            {{if .Stats}}
            // Initialize charts if stats are available
            initializeCharts();
            {{end}}
            
            // Setup sorting handlers
            setupSortHandlers();
        });
        
        function setupSortHandlers() {
            document.querySelectorAll('.sortable-header').forEach(header => {
                header.addEventListener('click', function(e) {
                    e.preventDefault();
                    
                    const sortType = this.dataset.sort;
                    
                    // Special handling for difficulty column (filtering)
                    if (sortType === 'difficulty') {
                        handleDifficultyFilter(this);
                        return;
                    }
                    
                    // Handle regular sorting
                    handleSort(this, sortType);
                });
            });
        }
        
        function handleDifficultyFilter(element) {
            // Cycle through difficulty filters
            const difficulties = ['all', 'basic', 'intermediate', 'hard', 'fun'];
            const currentIndex = difficulties.indexOf(currentDifficulty);
            const nextIndex = (currentIndex + 1) % difficulties.length;
            currentDifficulty = difficulties[nextIndex];
            
            // Update visual indicator
            updateDifficultyIndicator(element);
            
            // Make HTMX request with difficulty filter
            let url = '/leaderboard?sort=' + currentSort + '&order=' + currentOrder;
            if (currentDifficulty !== 'all') {
                url += '&difficulty=' + currentDifficulty;
            }
            
            htmx.ajax('GET', url, {
                target: '#leaderboard-content',
                swap: 'innerHTML'
            }).then(() => {
                setupSortHandlers();
            });
        }
        
        function handleSort(element, sortType) {
            // Add visual feedback
            element.classList.add('sorting-active');
            setTimeout(() => {
                element.classList.remove('sorting-active');
            }, 300);
            
            // Determine new sort order
            let newOrder;
            if (currentSort !== sortType) {
                newOrder = 'desc';
            } else {
                newOrder = currentOrder === 'desc' ? 'asc' : 'desc';
            }
            
            // Update current state
            currentSort = sortType;
            currentOrder = newOrder;
            
            // Make HTMX request
            let url = '/leaderboard?sort=' + sortType + '&order=' + newOrder;
            if (currentDifficulty !== 'all') {
                url += '&difficulty=' + currentDifficulty;
            }
            
            htmx.ajax('GET', url, {
                target: '#leaderboard-content',
                swap: 'innerHTML'
            }).then(() => {
                setupSortHandlers();
                updateSortIcons();
            });
        }
        
        function updateDifficultyIndicator(element) {
            const existing = element.querySelector('.difficulty-filter');
            if (existing) {
                existing.remove();
            }
            
            if (currentDifficulty !== 'all') {
                const indicator = document.createElement('span');
                indicator.className = 'difficulty-filter';
                
                const filterTexts = {
                    'basic': '🟢',
                    'intermediate': '🟡', 
                    'hard': '🔴',
                    'fun': '🎉'
                };
                
                indicator.textContent = filterTexts[currentDifficulty];
                element.appendChild(indicator);
            }
        }
        
        function updateSortIcons() {
            document.querySelectorAll('.sortable-header').forEach(header => {
                const sortType = header.dataset.sort;
                const icon = header.querySelector('.sort-icon');
                
                if (icon && sortType !== 'difficulty') {
                    if (currentSort !== sortType) {
                        icon.textContent = '↕️';
                    } else {
                        icon.textContent = currentOrder === 'desc' ? '↓' : '↑';
                    }
                }
                
                // Update active sort class
                if (sortType === currentSort) {
                    header.classList.add('active-sort');
                } else {
                    header.classList.remove('active-sort');
                }
            });
        }

        function initializeCharts() {
            // Get stats data from the template
            const stats = {{.Stats}};
            
            // Initialize Difficulty Distribution Chart
            initDifficultyChart(stats.by_difficulty);
            
            // Initialize Rule Progress Chart
            initProgressChart(stats.completion_rates);
        }
        
        function initDifficultyChart(difficultyData) {
            const ctx = document.getElementById('difficultyChart');
            if (!ctx) return;
            
            const difficulties = ['basic', 'intermediate', 'hard', 'fun'];
            const colors = ['#4ade80', '#facc15', '#f87171', '#a78bfa'];
            const icons = ['🟢', '🟡', '🔴', '🎉'];
            
            const data = difficulties.map(diff => difficultyData[diff] || 0);
            const labels = difficulties.map((diff, i) => icons[i] + ' ' + diff.charAt(0).toUpperCase() + diff.slice(1));
            
            new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels: labels,
                    datasets: [{
                        data: data,
                        backgroundColor: colors,
                        borderColor: colors.map(color => color + '80'),
                        borderWidth: 2,
                        hoverOffset: 4
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'bottom',
                            labels: {
                                padding: 20,
                                usePointStyle: true,
                                color: '#e2e8f0'
                            }
                        },
                        tooltip: {
                            callbacks: {
                                label: function(context) {
                                    const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                    const percentage = total > 0 ? ((context.parsed / total) * 100).toFixed(1) : 0;
                                    return context.label + ': ' + context.parsed + ' players (' + percentage + '%)';
                                }
                            }
                        }
                    }
                }
            });
        }
        
        function initProgressChart(completionData) {
            const ctx = document.getElementById('progressChart');
            if (!ctx) return;
            
            const milestones = ['rule_5', 'rule_10', 'rule_15', 'rule_20'];
            const labels = ['Rule 5+', 'Rule 10+', 'Rule 15+', 'Rule 20'];
            const data = milestones.map(milestone => completionData[milestone] || 0);
            
            new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: labels,
                    datasets: [{
                        label: 'Completion Rate (%)',
                        data: data,
                        backgroundColor: [
                            '#4ade8080',
                            '#facc1580', 
                            '#f8717180',
                            '#a78bfa80'
                        ],
                        borderColor: [
                            '#4ade80',
                            '#facc15',
                            '#f87171', 
                            '#a78bfa'
                        ],
                        borderWidth: 2,
                        borderRadius: 4,
                        borderSkipped: false,
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 100,
                            ticks: {
                                callback: function(value) {
                                    return value + '%';
                                },
                                color: '#e2e8f0'
                            },
                            grid: {
                                color: '#334155'
                            }
                        },
                        x: {
                            ticks: {
                                color: '#e2e8f0'
                            },
                            grid: {
                                color: '#334155'
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: false
                        },
                        tooltip: {
                            callbacks: {
                                label: function(context) {
                                    return context.parsed.y.toFixed(1) + '% of players reached this milestone';
                                }
                            }
                        }
                    }
                }
            });
        }
    </script>
</body>
</html>`

// leaderboardTableTemplate is the HTML template for just the table portion
const leaderboardTableTemplate = `{{define "leaderboard-table"}}
<div id="leaderboard-table">
    {{if .HasUsers}}
    <div class="table-header">
        <div>Rank</div>
        <div>Player</div>
        <div class="sortable-header {{if eq .SortBy "difficulty"}}active-sort{{end}}" 
             data-sort="difficulty">
            Difficulty<span class="sort-icon">🔄</span>
            <span class="sort-indicator htmx-indicator">↻</span>
        </div>
        <div class="sortable-header {{if eq .SortBy "rule"}}active-sort{{end}}" 
             data-sort="rule">
            Rules<span class="sort-icon">{{getSortIcon .SortBy "rule" .SortOrder}}</span>
            <span class="sort-indicator htmx-indicator">↻</span>
        </div>
        <div class="sortable-header {{if eq .SortBy "time"}}active-sort{{end}}" 
             data-sort="time">
            Time<span class="sort-icon">{{getSortIcon .SortBy "time" .SortOrder}}</span>
            <span class="sort-indicator htmx-indicator">↻</span>
        </div>
        <div class="sortable-header {{if eq .SortBy "joined"}}active-sort{{end}}" 
             data-sort="joined">
            Joined<span class="sort-icon">{{getSortIcon .SortBy "joined" .SortOrder}}</span>
            <span class="sort-indicator htmx-indicator">↻</span>
        </div>
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
    {{else}}
    <div class="empty-state">
        <h3>🎮 No Players Yet!</h3>
        <p>Be the first to join the Password Game challenge!</p>
        <a href="/" class="btn-primary" style="margin-top: 1rem; display: inline-block;">Start Playing</a>
    </div>
    {{end}}
</div>
{{end}}`
