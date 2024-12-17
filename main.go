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
			slog.Error("при работе произошла ошибка", "error", r)
			fmt.Println("Нажмите Enter для выхода...")
			fmt.Scanln() // Ждет нажатия Enter
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
		fmt.Print("Введите температуру в градусах кельвина: ")
		fmt.Scanln(&temperature)

		fmt.Print("Введите количество шагов: ")
		fmt.Scanln(&steps)
	}

	simulator := simulator.NewSimulator(cfg, temperature, steps)
	simulator.Simulate()

	if len(args) != 3 {
		// Ожидаем нажатия Enter для выхода
		fmt.Println("Нажмите Enter для выхода...")
		fmt.Scanln() // Ждет нажатия Enter
	}
}
