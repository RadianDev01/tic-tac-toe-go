# Tic Tac Toe (Go)

A simple command-line Tic Tac Toe game written in Go.

## Features

- Two-player gameplay
- Interactive CLI interface
- Input validation
- Play multiple games in a row

## Deployment Options

### Option 1: Local Deployment (Quickest)

#### Prerequisites
- Go 1.16 or higher installed

#### Steps
```bash
# Build the game
go build -o tictactoe tictactoe.go

# Run the game
./tictactoe
```

#### Windows
```bash
go build -o tictactoe.exe tictactoe.go
tictactoe.exe
```

### Option 2: Using Make (Recommended)

#### Prerequisites
- Go 1.16 or higher
- Make installed

#### Steps
```bash
# Build for your current platform
make build

# Run the game
make run

# Or build and run in one step
make

# Build for all platforms (Linux, macOS, Windows)
make build-all

# Clean build artifacts
make clean
```

### Option 3: Docker Deployment (Most Portable)

#### Prerequisites
- Docker installed

#### Steps
```bash
# Build the Docker image
docker build -t tictactoe-go .

# Run the game in a container
docker run -it tictactoe-go
```

This option is great for:
- Testing in an isolated environment
- Distributing to users without Go installed
- Ensuring consistent runtime environment

### Option 4: Direct Run (No Build)

For quick testing without building:
```bash
go run tictactoe.go
```

## How to Play

1. Start the game using one of the methods above
2. Players take turns entering their moves
3. Enter moves as "row column" (e.g., "1 2" for row 1, column 2)
4. Rows and columns are numbered 1-3
5. First player to get three in a row wins!
6. Choose to play again or quit after each game

## Game Example

```
     1   2   3
   +---+---+---+
 1 |   |   |   |
   +---+---+---+
 2 |   |   |   |
   +---+---+---+
 3 |   |   |   |
   +---+---+---+

Player X's turn
Enter your move (row col, e.g., '1 2'): 2 2
```

## Distribution

### Sharing with Users

After building, you can share the binary:

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o tictactoe-linux tictactoe.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o tictactoe-macos tictactoe.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o tictactoe-windows.exe tictactoe.go
```

Users can then run the appropriate binary for their platform without needing Go installed.

## Troubleshooting

**Issue**: `go: command not found`
**Solution**: Install Go from https://golang.org/dl/

**Issue**: Docker container exits immediately
**Solution**: Make sure to use `-it` flags for interactive mode: `docker run -it tictactoe-go`

**Issue**: Permission denied when running binary
**Solution**: Make the file executable: `chmod +x tictactoe`
