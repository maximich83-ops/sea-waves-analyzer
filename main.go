package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/waves"
	"github.com/maximich83-ops/sea-waves-analyzer/analyzer"
)

// Проміжна версія: введення як у ЛР2, але всі обчислення — у пакеті analyzer.
func main() {
	in := bufio.NewReader(os.Stdin)

	var count int
	fmt.Print("Введіть кількість значень (5-100): ")
	if _, err := fmt.Fscan(in, &count); err != nil {
		fmt.Fprintln(os.Stderr, "Помилка: кількість має бути цілим числом.")
		os.Exit(2)
	}
	if err := analyzer.ValidateCount(count); err != nil {
		fmt.Fprintln(os.Stderr, "Помилка:", err)
		os.Exit(2)
	}

	in.ReadString('\n') // залишок рядка після числа
	fmt.Print("Введіть назву морської акваторії: ")
	raw, _ := in.ReadString('\n')
	seaArea, err := analyzer.NormalizeSeaArea(raw)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Помилка:", err)
		os.Exit(2)
	}

	values, err := waves.Generate(count, seaArea)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Помилка генерації:", err)
		os.Exit(1)
	}
	fmt.Print("Введіть значення висоти хвилі: ")
	var limit float64
	if _, err := fmt.Fscan(in, &limit); err != nil {
		fmt.Fprintln(os.Stderr, "Помилка введення висоти:", err)
		os.Exit(2)
	}
	if err := analyzer.ValidateLimit(limit); err != nil {
		fmt.Fprintln(os.Stderr, "Помилка:", err)
		os.Exit(2)
	}

	series, err := analyzer.NewWaveSeries(seaArea, values)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Помилка даних:", err)
		os.Exit(1)
	}
	r := series.Analyze()
	if pos, ok := series.FirstAbove(limit); ok {
		fmt.Println("Номер першої хвилі:", pos)
	} else {
		fmt.Println("Таких хвиль немає")
	}

	fmt.Printf("\nАкваторія: %s\n", r.SeaArea)
	fmt.Println("Висота хвиль, м:", r.Heights)
	fmt.Printf("Мінімум: %.3f м\nМаксимум: %.3f м\nСума: %.3f м\nСереднє: %.3f м\n",
		r.Stats.Min, r.Stats.Max, r.Stats.Sum, r.Stats.Mean)
	fmt.Println("\nКатегорії:")
	for _, cc := range r.Categories {
		fmt.Printf("%s: %d\n", cc.Category, cc.Count)
	}
	fmt.Println("\nПовтори:")
	for _, rep := range r.Repeats {
		fmt.Printf("%.3f м — %d раз(и)\n", rep.Height, rep.Occurrences)
	}
	fmt.Println("Кількість однакових пар:", r.TotalPairs())
	if r.HasDangerous() {
		fmt.Println("\n!!! ПОПЕРЕДЖЕННЯ !!!\nВиявлено небезпечну хвилю понад 5 м.")
	}
}
