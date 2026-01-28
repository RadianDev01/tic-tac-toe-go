package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// User represents a player with their scores
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Scores    Scores    `json:"scores"`
	CreatedAt time.Time `json:"created_at"`
}

// Scores tracks wins, losses, and draws
type Scores struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

// Database holds all users
type Database struct {
	Users map[string]*User `json:"users"` // keyed by ID
	mu    sync.RWMutex
}

// Session stores active user sessions
type SessionStore struct {
	sessions map[string]string // token -> userID
	mu       sync.RWMutex
}

var (
	db       *Database
	sessions *SessionStore
	dbFile   = "users.json"
)

func main() {
	// Initialize database and sessions
	db = &Database{Users: make(map[string]*User)}
	sessions = &SessionStore{sessions: make(map[string]string)}

	// Load existing data
	loadDatabase()

	// API routes
	http.HandleFunc("/api/register", corsMiddleware(handleRegister))
	http.HandleFunc("/api/login", corsMiddleware(handleLogin))
	http.HandleFunc("/api/logout", corsMiddleware(handleLogout))
	http.HandleFunc("/api/user", corsMiddleware(handleGetUser))
	http.HandleFunc("/api/score", corsMiddleware(handleUpdateScore))
	http.HandleFunc("/api/leaderboard", corsMiddleware(handleLeaderboard))

	// Serve static files
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	port := "8080"
	fmt.Printf("Starting Tic Tac Toe web server on http://localhost:%s\n", port)
	fmt.Println("Open your browser and navigate to the URL above to play!")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// generateID creates a unique ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateToken creates a session token
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// loadDatabase reads users from JSON file
func loadDatabase() {
	data, err := os.ReadFile(dbFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("No existing database, starting fresh")
			return
		}
		log.Printf("Error reading database: %v", err)
		return
	}

	if err := json.Unmarshal(data, db); err != nil {
		log.Printf("Error parsing database: %v", err)
	}

	if db.Users == nil {
		db.Users = make(map[string]*User)
	}

	log.Printf("Loaded %d users from database", len(db.Users))
}

// saveDatabase writes users to JSON file
func saveDatabase() error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dbFile, data, 0644)
}

// findUserByUsername finds a user by username
func findUserByUsername(username string) *User {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, user := range db.Users {
		if user.Username == username {
			return user
		}
	}
	return nil
}

// getUserFromToken gets user from session token
func getUserFromToken(r *http.Request) *User {
	token := r.Header.Get("Authorization")
	if token == "" {
		return nil
	}

	sessions.mu.RLock()
	userID, exists := sessions.sessions[token]
	sessions.mu.RUnlock()

	if !exists {
		return nil
	}

	db.mu.RLock()
	user := db.Users[userID]
	db.mu.RUnlock()

	return user
}

// handleRegister creates a new user
func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || len(req.Username) < 2 || len(req.Username) > 20 {
		jsonError(w, "Username must be 2-20 characters", http.StatusBadRequest)
		return
	}

	// Check if username exists
	if findUserByUsername(req.Username) != nil {
		jsonError(w, "Username already taken", http.StatusConflict)
		return
	}

	// Create new user
	user := &User{
		ID:        generateID(),
		Username:  req.Username,
		Scores:    Scores{},
		CreatedAt: time.Now(),
	}

	db.mu.Lock()
	db.Users[user.ID] = user
	db.mu.Unlock()

	if err := saveDatabase(); err != nil {
		log.Printf("Error saving database: %v", err)
	}

	// Create session
	token := generateToken()
	sessions.mu.Lock()
	sessions.sessions[token] = user.ID
	sessions.mu.Unlock()

	jsonResponse(w, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// handleLogin logs in an existing user
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := findUserByUsername(req.Username)
	if user == nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	// Create session
	token := generateToken()
	sessions.mu.Lock()
	sessions.sessions[token] = user.ID
	sessions.mu.Unlock()

	jsonResponse(w, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// handleLogout logs out a user
func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token != "" {
		sessions.mu.Lock()
		delete(sessions.sessions, token)
		sessions.mu.Unlock()
	}

	jsonResponse(w, map[string]string{"status": "ok"})
}

// handleGetUser returns current user info
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromToken(r)
	if user == nil {
		jsonError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	jsonResponse(w, user)
}

// handleUpdateScore updates user's score
func handleUpdateScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUserFromToken(r)
	if user == nil {
		jsonError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var req struct {
		Result string `json:"result"` // "win", "loss", or "draw"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db.mu.Lock()
	switch req.Result {
	case "win":
		user.Scores.Wins++
	case "loss":
		user.Scores.Losses++
	case "draw":
		user.Scores.Draws++
	default:
		db.mu.Unlock()
		jsonError(w, "Invalid result type", http.StatusBadRequest)
		return
	}
	db.mu.Unlock()

	if err := saveDatabase(); err != nil {
		log.Printf("Error saving database: %v", err)
	}

	jsonResponse(w, user)
}

// handleLeaderboard returns top players
func handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db.mu.RLock()
	users := make([]*User, 0, len(db.Users))
	for _, user := range db.Users {
		users = append(users, user)
	}
	db.mu.RUnlock()

	// Sort by wins (simple bubble sort for small lists)
	for i := 0; i < len(users)-1; i++ {
		for j := 0; j < len(users)-i-1; j++ {
			if users[j].Scores.Wins < users[j+1].Scores.Wins {
				users[j], users[j+1] = users[j+1], users[j]
			}
		}
	}

	// Return top 10
	limit := 10
	if len(users) < limit {
		limit = len(users)
	}

	jsonResponse(w, users[:limit])
}

// jsonResponse sends a JSON response
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// jsonError sends a JSON error response
func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
