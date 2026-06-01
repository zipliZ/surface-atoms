package main

import (
	"fmt"
	"log"
	"log/slog"
	"main/configs"
	"main/internal/simulation"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("An error occurred during execution", "error", r)
			fmt.Println(string(debug.Stack()))
			fmt.Println("Press Enter to exit...")
			fmt.Scanln() // Waits for Enter key press
		}
	}()

	cfg, err := configs.New()
	if err != nil {
		log.Panic(err)
	}

	var temperature int
	var simulationTime float64

	args := os.Args
	if len(args) == 3 {
		temperature, err = strconv.Atoi(args[1])
		if err != nil {
			log.Fatal(err) // nolint
		}
		args[2] = strings.ReplaceAll(args[2], "_", "")
		simulationTime, err = strconv.ParseFloat(args[2], 64)
		if err != nil {
			log.Fatal(err)
		}
		slog.SetLogLoggerLevel(slog.LevelError)
	} else {
		fmt.Print("Enter the temperature in Kelvin: ")
		fmt.Scanln(&temperature)

		fmt.Print("Enter simulation time: ")
		fmt.Scanln(&simulationTime)
	}

	simulator, err := simulation.NewSimulator(cfg, temperature, simulationTime)
	if err != nil {
		log.Fatal(err)
	}
	if err = simulator.Simulate(); err != nil {
		log.Fatal(err)
	}

	if len(args) != 3 {
		// Wait for Enter key press before exiting
		fmt.Println("Press Enter to exit...")
		fmt.Scanln() // Waits for Enter key press
	}
}
