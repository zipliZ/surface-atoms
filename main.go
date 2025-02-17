package main

import (
	"fmt"
	"log"
	"log/slog"
	"main/internal/config"
	"main/internal/simulator"
	"os"
	"strconv"
	"strings"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("An error occurred during execution", "error", r)
			fmt.Println("Press Enter to exit...")
			fmt.Scanln() // Waits for Enter key press
		}
	}()

	cfg, err := config.New()
	if err != nil {
		log.Panic(err)
	}

	var temperature int
	var steps int

	args := os.Args
	if len(args) == 3 {
		temperature, err = strconv.Atoi(args[1])
		if err != nil {
			log.Fatal(err)
		}
		args[2] = strings.ReplaceAll(args[2], "_", "")
		steps, err = strconv.Atoi(args[2])
		if err != nil {
			log.Fatal(err)
		}
		slog.SetLogLoggerLevel(slog.LevelError)
	} else {
		fmt.Print("Enter the temperature in Kelvin: ")
		fmt.Scanln(&temperature)

		fmt.Print("Enter the number of steps: ")
		fmt.Scanln(&steps)
	}

	simulator := simulator.NewSimulator(cfg, temperature, steps)
	simulator.Simulate()

	if len(args) != 3 {
		// Wait for Enter key press before exiting
		fmt.Println("Press Enter to exit...")
		fmt.Scanln() // Waits for Enter key press
	}
}
