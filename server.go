package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

// GameRoom represents an online multiplayer game
type GameRoom struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"` // 6-char join code
	BoardSize   int       `json:"board_size"`
	Board       []string  `json:"board"`
	PlayerX     *User     `json:"player_x"`
	PlayerO     *User     `json:"player_o"`
	CurrentTurn string    `json:"current_turn"` // "X" or "O"
	Status      string    `json:"status"`       // "waiting", "playing", "finished"
	Winner      string    `json:"winner"`       // "X", "O", "draw", or ""
	WinningLine []int     `json:"winning_line"` // indices of winning cells
	LastMove    int       `json:"last_move"`    // index of last move
	ShowEmote   bool      `json:"show_emote"`   // whether to show emote
	EmoteType   string    `json:"emote_type"`   // type of emote (e.g., "deal_with_it")
	EmoteBy     string    `json:"emote_by"`     // username who triggered it
	EmoteAt     time.Time `json:"emote_at"`     // when emote was triggered
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GameStore manages active game rooms
type GameStore struct {
	rooms map[string]*GameRoom // keyed by room ID
	codes map[string]string    // code -> room ID
	mu    sync.RWMutex
}

var (
	db       *Database
	sessions *SessionStore
	games    *GameStore
	dbFile   = "users.json"
)

func main() {
	// Initialize database, sessions, and games
	db = &Database{Users: make(map[string]*User)}
	sessions = &SessionStore{sessions: make(map[string]string)}
	games = &GameStore{
		rooms: make(map[string]*GameRoom),
		codes: make(map[string]string),
	}

	// Load existing data
	loadDatabase()

	// Start cleanup routine for old games
	go cleanupOldGames()

	// API routes - User management
	http.HandleFunc("/api/register", corsMiddleware(handleRegister))
	http.HandleFunc("/api/login", corsMiddleware(handleLogin))
	http.HandleFunc("/api/logout", corsMiddleware(handleLogout))
	http.HandleFunc("/api/user", corsMiddleware(handleGetUser))
	http.HandleFunc("/api/score", corsMiddleware(handleUpdateScore))
	http.HandleFunc("/api/leaderboard", corsMiddleware(handleLeaderboard))

	// API routes - Multiplayer games
	http.HandleFunc("/api/game/create", corsMiddleware(handleCreateGame))
	http.HandleFunc("/api/game/join", corsMiddleware(handleJoinGame))
	http.HandleFunc("/api/game/state", corsMiddleware(handleGameState))
	http.HandleFunc("/api/game/move", corsMiddleware(handleGameMove))
	http.HandleFunc("/api/game/leave", corsMiddleware(handleLeaveGame))
	http.HandleFunc("/api/game/emote", corsMiddleware(handleGameEmote))

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

// generateGameCode creates a 6-character game code
func generateGameCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed confusing chars
	bytes := make([]byte, 6)
	rand.Read(bytes)
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[int(bytes[i])%len(chars)]
	}
	return string(code)
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

// cleanupOldGames removes games older than 1 hour
func cleanupOldGames() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		games.mu.Lock()
		now := time.Now()
		for id, room := range games.rooms {
			if now.Sub(room.UpdatedAt) > time.Hour {
				delete(games.codes, room.Code)
				delete(games.rooms, id)
				log.Printf("Cleaned up old game room: %s", room.Code)
			}
		}
		games.mu.Unlock()
	}
}

// generateWinningConditions creates all winning line combinations
func generateWinningConditions(size int) [][]int {
	winLen := 3
	if size == 5 {
		winLen = 4
	}

	var conditions [][]int

	// Rows
	for row := 0; row < size; row++ {
		for startCol := 0; startCol <= size-winLen; startCol++ {
			condition := make([]int, winLen)
			for i := 0; i < winLen; i++ {
				condition[i] = row*size + startCol + i
			}
			conditions = append(conditions, condition)
		}
	}

	// Columns
	for col := 0; col < size; col++ {
		for startRow := 0; startRow <= size-winLen; startRow++ {
			condition := make([]int, winLen)
			for i := 0; i < winLen; i++ {
				condition[i] = (startRow+i)*size + col
			}
			conditions = append(conditions, condition)
		}
	}

	// Diagonals (top-left to bottom-right)
	for row := 0; row <= size-winLen; row++ {
		for col := 0; col <= size-winLen; col++ {
			condition := make([]int, winLen)
			for i := 0; i < winLen; i++ {
				condition[i] = (row+i)*size + col + i
			}
			conditions = append(conditions, condition)
		}
	}

	// Diagonals (top-right to bottom-left)
	for row := 0; row <= size-winLen; row++ {
		for col := winLen - 1; col < size; col++ {
			condition := make([]int, winLen)
			for i := 0; i < winLen; i++ {
				condition[i] = (row+i)*size + col - i
			}
			conditions = append(conditions, condition)
		}
	}

	return conditions
}

// checkWinner checks if there's a winner
func checkWinner(board []string, size int) (string, []int) {
	conditions := generateWinningConditions(size)

	for _, condition := range conditions {
		first := board[condition[0]]
		if first == "" {
			continue
		}

		won := true
		for _, idx := range condition {
			if board[idx] != first {
				won = false
				break
			}
		}

		if won {
			return first, condition
		}
	}

	return "", nil
}

// checkDraw checks if the game is a draw
func checkDraw(board []string) bool {
	for _, cell := range board {
		if cell == "" {
			return false
		}
	}
	return true
}

// ==================== User Management Handlers ====================

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

// ==================== Game Room Handlers ====================

// handleCreateGame creates a new game room
func handleCreateGame(w http.ResponseWriter, r *http.Request) {
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
		BoardSize int `json:"board_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.BoardSize != 3 && req.BoardSize != 5 {
		req.BoardSize = 3
	}

	// Generate unique code
	var code string
	games.mu.Lock()
	for {
		code = generateGameCode()
		if _, exists := games.codes[code]; !exists {
			break
		}
	}

	room := &GameRoom{
		ID:          generateID(),
		Code:        code,
		BoardSize:   req.BoardSize,
		Board:       make([]string, req.BoardSize*req.BoardSize),
		PlayerX:     user,
		PlayerO:     nil,
		CurrentTurn: "X",
		Status:      "waiting",
		Winner:      "",
		WinningLine: nil,
		LastMove:    -1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	games.rooms[room.ID] = room
	games.codes[code] = room.ID
	games.mu.Unlock()

	log.Printf("Game created: %s by %s", code, user.Username)

	jsonResponse(w, room)
}

// handleJoinGame joins an existing game room
func handleJoinGame(w http.ResponseWriter, r *http.Request) {
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
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(req.Code))

	games.mu.Lock()
	roomID, exists := games.codes[code]
	if !exists {
		games.mu.Unlock()
		jsonError(w, "Game not found", http.StatusNotFound)
		return
	}

	room := games.rooms[roomID]
	if room == nil {
		games.mu.Unlock()
		jsonError(w, "Game not found", http.StatusNotFound)
		return
	}

	// Check if user is already in this game
	if room.PlayerX != nil && room.PlayerX.ID == user.ID {
		games.mu.Unlock()
		jsonResponse(w, room)
		return
	}

	if room.PlayerO != nil && room.PlayerO.ID == user.ID {
		games.mu.Unlock()
		jsonResponse(w, room)
		return
	}

	// Check if game is full
	if room.PlayerO != nil {
		games.mu.Unlock()
		jsonError(w, "Game is full", http.StatusConflict)
		return
	}

	// Join as player O
	room.PlayerO = user
	room.Status = "playing"
	room.UpdatedAt = time.Now()
	games.mu.Unlock()

	log.Printf("Game %s: %s joined as O", code, user.Username)

	jsonResponse(w, room)
}

// handleGameState returns current game state
func handleGameState(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		jsonError(w, "Room ID required", http.StatusBadRequest)
		return
	}

	games.mu.Lock()
	room := games.rooms[roomID]
	if room != nil {
		// Auto-clear emote after 3 seconds
		if room.ShowEmote && time.Since(room.EmoteAt) > 3*time.Second {
			room.ShowEmote = false
			room.EmoteType = ""
			room.EmoteBy = ""
		}
	}
	games.mu.Unlock()

	if room == nil {
		jsonError(w, "Game not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, room)
}

// handleGameMove processes a player's move
func handleGameMove(w http.ResponseWriter, r *http.Request) {
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
		RoomID string `json:"room_id"`
		Index  int    `json:"index"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	games.mu.Lock()
	room := games.rooms[req.RoomID]
	if room == nil {
		games.mu.Unlock()
		jsonError(w, "Game not found", http.StatusNotFound)
		return
	}

	// Verify game is in progress
	if room.Status != "playing" {
		games.mu.Unlock()
		jsonError(w, "Game is not in progress", http.StatusBadRequest)
		return
	}

	// Verify it's this player's turn
	var playerSymbol string
	if room.PlayerX != nil && room.PlayerX.ID == user.ID {
		playerSymbol = "X"
	} else if room.PlayerO != nil && room.PlayerO.ID == user.ID {
		playerSymbol = "O"
	} else {
		games.mu.Unlock()
		jsonError(w, "You are not in this game", http.StatusForbidden)
		return
	}

	if room.CurrentTurn != playerSymbol {
		games.mu.Unlock()
		jsonError(w, "Not your turn", http.StatusBadRequest)
		return
	}

	// Verify move is valid
	if req.Index < 0 || req.Index >= len(room.Board) {
		games.mu.Unlock()
		jsonError(w, "Invalid move position", http.StatusBadRequest)
		return
	}

	if room.Board[req.Index] != "" {
		games.mu.Unlock()
		jsonError(w, "Cell already taken", http.StatusBadRequest)
		return
	}

	// Make the move
	room.Board[req.Index] = playerSymbol
	room.LastMove = req.Index
	room.UpdatedAt = time.Now()

	// Check for winner
	winner, winningLine := checkWinner(room.Board, room.BoardSize)
	if winner != "" {
		room.Winner = winner
		room.WinningLine = winningLine
		room.Status = "finished"

		// Update scores
		if winner == "X" && room.PlayerX != nil {
			room.PlayerX.Scores.Wins++
			if room.PlayerO != nil {
				room.PlayerO.Scores.Losses++
			}
		} else if winner == "O" && room.PlayerO != nil {
			room.PlayerO.Scores.Wins++
			if room.PlayerX != nil {
				room.PlayerX.Scores.Losses++
			}
		}
		saveDatabase()
	} else if checkDraw(room.Board) {
		room.Winner = "draw"
		room.Status = "finished"

		// Update scores
		if room.PlayerX != nil {
			room.PlayerX.Scores.Draws++
		}
		if room.PlayerO != nil {
			room.PlayerO.Scores.Draws++
		}
		saveDatabase()
	} else {
		// Switch turns
		if room.CurrentTurn == "X" {
			room.CurrentTurn = "O"
		} else {
			room.CurrentTurn = "X"
		}
	}

	games.mu.Unlock()

	jsonResponse(w, room)
}

// handleLeaveGame removes a player from a game
func handleLeaveGame(w http.ResponseWriter, r *http.Request) {
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
		RoomID string `json:"room_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	games.mu.Lock()
	room := games.rooms[req.RoomID]
	if room == nil {
		games.mu.Unlock()
		jsonResponse(w, map[string]string{"status": "ok"})
		return
	}

	// If game is waiting or finished, just delete it
	if room.Status == "waiting" || room.Status == "finished" {
		delete(games.codes, room.Code)
		delete(games.rooms, room.ID)
		games.mu.Unlock()
		jsonResponse(w, map[string]string{"status": "ok"})
		return
	}

	// If game is in progress, the leaving player forfeits
	if room.PlayerX != nil && room.PlayerX.ID == user.ID {
		room.Winner = "O"
		room.Status = "finished"
		if room.PlayerO != nil {
			room.PlayerO.Scores.Wins++
		}
		room.PlayerX.Scores.Losses++
		saveDatabase()
	} else if room.PlayerO != nil && room.PlayerO.ID == user.ID {
		room.Winner = "X"
		room.Status = "finished"
		if room.PlayerX != nil {
			room.PlayerX.Scores.Wins++
		}
		room.PlayerO.Scores.Losses++
		saveDatabase()
	}

	games.mu.Unlock()
	jsonResponse(w, map[string]string{"status": "ok"})
}

// handleGameEmote triggers an emote for both players to see
func handleGameEmote(w http.ResponseWriter, r *http.Request) {
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
		RoomID    string `json:"room_id"`
		EmoteType string `json:"emote_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	games.mu.Lock()
	room := games.rooms[req.RoomID]
	if room == nil {
		games.mu.Unlock()
		jsonError(w, "Game not found", http.StatusNotFound)
		return
	}

	// Verify user is in this game
	isInGame := (room.PlayerX != nil && room.PlayerX.ID == user.ID) ||
		(room.PlayerO != nil && room.PlayerO.ID == user.ID)
	if !isInGame {
		games.mu.Unlock()
		jsonError(w, "You are not in this game", http.StatusForbidden)
		return
	}

	// Set the emote
	room.ShowEmote = true
	room.EmoteType = req.EmoteType
	room.EmoteBy = user.Username
	room.EmoteAt = time.Now()
	room.UpdatedAt = time.Now()

	games.mu.Unlock()

	log.Printf("Game %s: %s triggered emote %s", room.Code, user.Username, req.EmoteType)

	jsonResponse(w, room)
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
