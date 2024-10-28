package main

import (
	"fmt"
	"main/internal/simulator"
)

func main() {
	var temperature int
	var steps int
	var n int

	fmt.Print("Введите температуру в градусах цельсия: ")
	fmt.Scanln(&temperature)

	fmt.Print("Введите количество шагов: ")
	fmt.Scanln(&steps)

	fmt.Print("Введите размер матрицы: ")
	fmt.Scanln(&n)

	simulator := simulator.NewSimulator(temperature+253, steps, n, n)
	simulator.Simulate()

	// Ожидаем нажатия Enter для выхода
	fmt.Println("Нажмите Enter для выхода...")
	fmt.Scanln() // Ждет нажатия Enter
}
