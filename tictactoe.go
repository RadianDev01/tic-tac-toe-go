package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Board [3][3]string

const (
	Empty  = " "
	PlayerX = "X"
	PlayerO = "O"
)

func main() {
	fmt.Println("Welcome to Tic Tac Toe!")
	fmt.Println("======================")
	
	for {
		playGame()
		
		fmt.Print("\nPlay again? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		
		if input != "y" && input != "yes" {
			fmt.Println("Thanks for playing!")
			break
		}
		fmt.Println()
	}
}

func playGame() {
	board := initBoard()
	currentPlayer := PlayerX
	moveCount := 0
	
	for {
		printBoard(board)
		fmt.Printf("\nPlayer %s's turn\n", currentPlayer)
		
		row, col := getMove(board)
		board[row][col] = currentPlayer
		moveCount++
		
		if checkWinner(board, currentPlayer) {
			printBoard(board)
			fmt.Printf("\n🎉 Player %s wins!\n", currentPlayer)
			break
		}
		
		if moveCount == 9 {
			printBoard(board)
			fmt.Println("\n🤝 It's a draw!")
			break
		}
		
		if currentPlayer == PlayerX {
			currentPlayer = PlayerO
		} else {
			currentPlayer = PlayerX
		}
	}
}

func initBoard() Board {
	var board Board
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			board[i][j] = Empty
		}
	}
	return board
}

func printBoard(board Board) {
	fmt.Println("\n     1   2   3")
	fmt.Println("   +---+---+---+")
	for i := 0; i < 3; i++ {
		fmt.Printf(" %d | %s | %s | %s |\n", i+1, board[i][0], board[i][1], board[i][2])
		fmt.Println("   +---+---+---+")
	}
}

func getMove(board Board) (int, int) {
	reader := bufio.NewReader(os.Stdin)
	
	for {
		fmt.Print("Enter your move (row col, e.g., '1 2'): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		parts := strings.Fields(input)
		if len(parts) != 2 {
			fmt.Println("Invalid input. Please enter row and column separated by space.")
			continue
		}
		
		row, err1 := strconv.Atoi(parts[0])
		col, err2 := strconv.Atoi(parts[1])
		
		if err1 != nil || err2 != nil {
			fmt.Println("Invalid input. Please enter numbers only.")
			continue
		}
		
		row--
		col--
		
		if row < 0 || row > 2 || col < 0 || col > 2 {
			fmt.Println("Invalid position. Row and column must be between 1 and 3.")
			continue
		}
		
		if board[row][col] != Empty {
			fmt.Println("That position is already taken. Try again.")
			continue
		}
		
		return row, col
	}
}

func checkWinner(board Board, player string) bool {
	// Check rows
	for i := 0; i < 3; i++ {
		if board[i][0] == player && board[i][1] == player && board[i][2] == player {
			return true
		}
	}
	
	// Check columns
	for j := 0; j < 3; j++ {
		if board[0][j] == player && board[1][j] == player && board[2][j] == player {
			return true
		}
	}
	
	// Check diagonals
	if board[0][0] == player && board[1][1] == player && board[2][2] == player {
		return true
	}
	if board[0][2] == player && board[1][1] == player && board[2][0] == player {
		return true
	}
	
	return false
}
