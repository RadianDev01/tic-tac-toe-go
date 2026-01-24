# Tic Tac Toe - Web Version

Play Tic Tac Toe in your browser!

## Features

- Beautiful, responsive UI with gradient styling
- Score tracking (persisted in browser localStorage)
- Winning cell highlighting
- Smooth animations and transitions
- Player turn indicator
- New game button to reset the board

## How to Run

1. Start the web server:
   ```bash
   go run server.go
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

3. Start playing! Click on any cell to make your move.

## Game Rules

- Players take turns placing X and O on the 3x3 grid
- The first player to get 3 of their marks in a row (horizontally, vertically, or diagonally) wins
- If all 9 cells are filled without a winner, the game is a draw
- Scores are tracked across games and saved in your browser

## Files

- `server.go` - Simple HTTP server to serve the game
- `index.html` - Complete game UI with embedded CSS and JavaScript
- `tictactoe.go` - Original command-line version (still available)
