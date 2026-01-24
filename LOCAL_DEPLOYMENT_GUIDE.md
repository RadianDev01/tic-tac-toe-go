# Local Deployment Guide - Tic Tac Toe Go

This guide will walk you through deploying and running the Tic Tac Toe game on your local machine, step by step.

## Table of Contents

1. [Prerequisites Check](#prerequisites-check)
2. [Method 1: Quick Run (Easiest)](#method-1-quick-run-easiest)
3. [Method 2: Build and Run](#method-2-build-and-run)
4. [Method 3: Using Make (Recommended)](#method-3-using-make-recommended)
5. [Method 4: Docker (Most Portable)](#method-4-docker-most-portable)
6. [Verification Steps](#verification-steps)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites Check

Before starting, let's make sure you have the necessary tools installed.

### Step 1: Check if Go is installed

Open your terminal and run:

```bash
go version
```

**Expected output:**
```
go version go1.21.x (or higher)
```

**If you see an error:**
- **Linux/macOS:** Install Go from https://golang.org/dl/
  ```bash
  # For Ubuntu/Debian
  sudo apt update
  sudo apt install golang-go

  # For macOS with Homebrew
  brew install go
  ```

- **Windows:** Download the installer from https://golang.org/dl/ and run it

### Step 2: Verify you're in the project directory

```bash
pwd
```

**Expected output:** Should show the path to `tic-tac-toe-go` directory

If not, navigate to it:
```bash
cd /path/to/tic-tac-toe-go
```

### Step 3: List the project files

```bash
ls -la
```

**You should see:**
- `tictactoe.go` - The main game file
- `Dockerfile` - Docker configuration
- `Makefile` - Build automation
- `README.md` - Project documentation

---

## Method 1: Quick Run (Easiest)

This method runs the game directly without building a binary. Perfect for quick testing!

### Step 1: Run the game

```bash
go run tictactoe.go
```

### Step 2: Play the game

You should see:
```
Welcome to Tic Tac Toe!
======================

     1   2   3
   +---+---+---+
 1 |   |   |   |
   +---+---+---+
 2 |   |   |   |
   +---+---+---+
 3 |   |   |   |
   +---+---+---+

Player X's turn
Enter your move (row col, e.g., '1 2'):
```

### Step 3: Make a move

Type a row and column number separated by a space:
```
2 2
```

This places your mark in the center of the board.

### Step 4: Exit the game

After a game ends, you'll be asked:
```
Play again? (y/n):
```

Type `n` and press Enter to quit.

**Pros:** No build step required, immediate testing
**Cons:** Slower startup, requires Go installed on target machine

---

## Method 2: Build and Run

This method creates a compiled binary that runs faster and can be distributed.

### Step 1: Build the game

```bash
go build -o tictactoe tictactoe.go
```

**What this does:** Compiles the Go code into an executable binary named `tictactoe`

**Expected output:** (command completes silently if successful)

### Step 2: Verify the binary was created

```bash
ls -lh tictactoe
```

**Expected output:**
```
-rwxr-xr-x 1 user user 2.1M Jan 24 12:00 tictactoe
```

### Step 3: Run the binary

**Linux/macOS:**
```bash
./tictactoe
```

**Windows:**
```bash
tictactoe.exe
```

**If you get "Permission denied" error:**
```bash
chmod +x tictactoe
./tictactoe
```

### Step 4: Play the game

The game interface is the same as Method 1.

**Pros:** Faster execution, can share binary with others
**Cons:** Need to rebuild after code changes

---

## Method 3: Using Make (Recommended)

Make simplifies the build and run process with simple commands.

### Step 1: Check if Make is installed

```bash
make --version
```

**If not installed:**
- **Linux:** `sudo apt install build-essential`
- **macOS:** `xcode-select --install`
- **Windows:** Install via [chocolatey](https://chocolatey.org/): `choco install make`

### Step 2: View available Make commands

```bash
make help
```

Or check the Makefile:
```bash
cat Makefile
```

### Available Commands:

#### Build the game
```bash
make build
```

**Output:**
```
Building tictactoe...
go build -o tictactoe tictactoe.go
Build complete!
```

#### Run the game
```bash
make run
```

**Output:**
```
Starting Tic Tac Toe...
./tictactoe
```

#### Build and run in one command
```bash
make
```

This runs the default target which builds and then runs the game.

#### Build for all platforms
```bash
make build-all
```

**What this does:** Creates binaries for:
- Linux (x86-64)
- macOS (Intel x86-64)
- macOS (Apple Silicon ARM64)
- Windows (x86-64)

**Output location:** `dist/` directory

**Files created:**
```
dist/tictactoe-linux-amd64
dist/tictactoe-darwin-amd64
dist/tictactoe-darwin-arm64
dist/tictactoe-windows-amd64.exe
```

#### Quick test run
```bash
make test-run
```

Runs the game without building a binary (same as `go run`).

#### Clean build artifacts
```bash
make clean
```

**What this does:** Removes compiled binaries and the `dist/` directory

**Pros:** Simple commands, automated workflow, cross-platform builds
**Cons:** Requires Make installed

---

## Method 4: Docker (Most Portable)

Docker ensures the game runs the same way on any system.

### Step 1: Check if Docker is installed

```bash
docker --version
```

**Expected output:**
```
Docker version 24.x.x, build xxxxx
```

**If not installed:** Visit https://docs.docker.com/get-docker/

### Step 2: Verify Docker is running

```bash
docker ps
```

**Expected output:** A table showing running containers (may be empty)

**If you see an error:**
- **Linux:** Start Docker: `sudo systemctl start docker`
- **macOS/Windows:** Start Docker Desktop application

### Step 3: Build the Docker image

```bash
docker build -t tictactoe-go .
```

**What this does:**
1. Uses the Dockerfile to create an image
2. Tags it as `tictactoe-go`
3. Compiles the Go code inside the container

**Expected output:**
```
[+] Building 5.2s (12/12) FINISHED
...
=> => naming to docker.io/library/tictactoe-go
```

**Alternative using Make:**
```bash
make docker-build
```

### Step 4: Verify the image was created

```bash
docker images | grep tictactoe-go
```

**Expected output:**
```
tictactoe-go   latest   abc123def456   2 minutes ago   10.5MB
```

### Step 5: Run the game in a container

```bash
docker run -it tictactoe-go
```

**Important flags:**
- `-i` = Interactive mode (allows input)
- `-t` = Allocates a pseudo-TTY (terminal)
- **Both are required** for the game to work properly

**Alternative using Make:**
```bash
make docker-run
```

### Step 6: Play the game

The game interface is the same as other methods.

### Step 7: Exit Docker container

After quitting the game, you'll automatically exit the container.

To manually exit: Press `Ctrl+D` or type `exit`

**Pros:** Consistent environment, no local Go installation needed, easy distribution
**Cons:** Requires Docker installed, slight overhead

---

## Verification Steps

After deployment, verify everything works correctly:

### 1. Game Starts Successfully
- You should see the welcome message
- The board displays correctly
- No error messages appear

### 2. Input Validation Works
Test invalid inputs:
```
Enter your move (row col, e.g., '1 2'): abc
Invalid input. Please enter numbers only.

Enter your move (row col, e.g., '1 2'): 5 5
Invalid position. Row and column must be between 1 and 3.

Enter your move (row col, e.g., '1 2'): 1 1
(Marks position)

Enter your move (row col, e.g., '1 2'): 1 1
That position is already taken. Try again.
```

### 3. Win Detection Works
Play a game and get three in a row. You should see:
```
🎉 Player X wins!
```

### 4. Draw Detection Works
Fill the board without a winner. You should see:
```
🤝 It's a draw!
```

### 5. Replay Feature Works
After a game ends:
```
Play again? (y/n): y
```
Should start a new game with a fresh board.

---

## Troubleshooting

### Problem: `go: command not found`

**Solution:**
1. Install Go from https://golang.org/dl/
2. Verify installation: `go version`
3. If still not found, add Go to your PATH:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH=$PATH:/usr/local/go/bin
   source ~/.bashrc  # or ~/.zshrc
   ```

### Problem: `./tictactoe: Permission denied`

**Solution:**
```bash
chmod +x tictactoe
./tictactoe
```

### Problem: Docker container exits immediately

**Solution:**
Must use `-it` flags together:
```bash
docker run -it tictactoe-go
```

NOT:
```bash
docker run tictactoe-go  # ❌ Wrong
docker run -i tictactoe-go  # ❌ Missing -t
docker run -t tictactoe-go  # ❌ Missing -i
```

### Problem: `Cannot connect to Docker daemon`

**Solution:**
1. **Linux:** Start Docker service
   ```bash
   sudo systemctl start docker
   ```

2. **macOS/Windows:** Open Docker Desktop application

3. Verify Docker is running:
   ```bash
   docker ps
   ```

### Problem: `make: command not found`

**Solution:**
- **Linux:** `sudo apt install build-essential`
- **macOS:** `xcode-select --install`
- **Windows:** Use Git Bash or install Make via Chocolatey

### Problem: Game displays incorrectly (missing characters/emojis)

**Solution:**
1. Ensure terminal supports UTF-8:
   ```bash
   echo $LANG
   ```
   Should show `en_US.UTF-8` or similar

2. Set UTF-8 if needed:
   ```bash
   export LANG=en_US.UTF-8
   ```

3. Use a modern terminal emulator (avoid old Windows CMD)

### Problem: Build fails with module errors

**Solution:**
This is a simple single-file Go program without modules. If you see module errors:
```bash
# Initialize module (if needed)
go mod init tictactoe

# Then build
go build tictactoe.go
```

### Problem: Port already in use (for future web versions)

**Solution:**
This CLI game doesn't use ports, but if you encounter this in future versions:
```bash
# Find process using port
lsof -i :8080

# Kill the process
kill -9 <PID>
```

---

## Quick Reference

| Method | Command | Prerequisites | Best For |
|--------|---------|---------------|----------|
| Quick Run | `go run tictactoe.go` | Go | Testing/Development |
| Build & Run | `go build -o tictactoe tictactoe.go && ./tictactoe` | Go | General Use |
| Make | `make` or `make build && make run` | Go, Make | Developers |
| Docker | `docker build -t tictactoe-go . && docker run -it tictactoe-go` | Docker | Distribution/Isolation |

---

## Next Steps

After successfully deploying locally, you might want to:

1. **Share with friends:** Use `make build-all` to create binaries for all platforms
2. **Modify the game:** Edit `tictactoe.go` and rebuild
3. **Create a web version:** Consider adding a web interface
4. **Deploy to cloud:** Use Docker image on cloud platforms
5. **Add AI opponent:** Enhance the game with computer player

---

## Need Help?

If you encounter issues not covered here:

1. Check the main [README.md](README.md) for additional information
2. Verify your Go version is 1.16 or higher: `go version`
3. Ensure all files are present: `ls -la`
4. Check file permissions: `ls -l tictactoe.go`

Happy gaming! 🎮
