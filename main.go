package main

import (
	"fmt"
	"log"
	"log/slog"
	"main/internal/config"
	"main/internal/simulator"
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
		log.Fatal(err)
	}

	var temperature int
	var steps int

	fmt.Print("Введите температуру в градусах кельвина: ")
	fmt.Scanln(&temperature)

	fmt.Print("Введите количество шагов: ")
	fmt.Scanln(&steps)

	simulator := simulator.NewSimulator(cfg, temperature, steps)
	simulator.Simulate()

	// Ожидаем нажатия Enter для выхода
	fmt.Println("Нажмите Enter для выхода...")
	fmt.Scanln() // Ждет нажатия Enter
}
