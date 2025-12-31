package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/nahtao97/scribble/internal/db"
	"github.com/nahtao97/scribble/internal/services"
)

func main() {
	fmt.Println("Starting daily challenge selection...")

	// Initialize database connection
	database, err := db.NewDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Create daily challenge service
	challengeService := services.NewDailyChallengeService(database.GetConnection())

	// Select next daily challenge
	challenge, err := challengeService.SelectNextChallenge()
	if err != nil {
		if errors.Is(err, services.ErrChallengeExists) {
			fmt.Printf("Daily challenge already exists for today: Problem ID %d\n", challenge.ProblemID)
			os.Exit(0)
		}
		if errors.Is(err, services.ErrNoProblems) {
			fmt.Fprintln(os.Stderr, "No problems available in database")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Failed to select daily challenge: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created daily challenge:\n")
	fmt.Printf("  Challenge ID: %d\n", challenge.ID)
	fmt.Printf("  Problem ID:   %d\n", challenge.ProblemID)
	fmt.Printf("  Problem:      %s\n", challenge.Problem.Title)
	fmt.Printf("  Difficulty:   %s\n", challenge.Problem.Difficulty)
	fmt.Printf("  Date:         %s\n", challenge.ChallengeDate.Format("2006-01-02"))
}
