package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/waves"
)

func category(h float64) string {
	switch {
	case h <= 0.5:
		return "спокійно"
	case h <= 2.0:
		return "невеликі хвилі"
	case h <= 5.0:
		return "значні"
	default:
		return "небезпечні"
	}
}

func main() {
	in := bufio.NewReader(os.Stdin)

	var count int

	fmt.Print("Введіть кількість значень (5-100): ")
	if _, err := fmt.Fscan(in, &count); err != nil {
		fmt.Println("Помилка введення кількості:", err)
		return
	}

	if count < 5 || count > 100 {
		fmt.Println("Помилка: кількість повинна бути від 5 до 100.")
		return
	}

	in.ReadString('\n')

	fmt.Print("Введіть назву морської акваторії: ")
	seaArea, err := in.ReadString('\n')
	if err != nil && len(seaArea) == 0 {
		fmt.Println("Помилка введення акваторії:", err)
		return
	}

	seaArea = strings.TrimSpace(seaArea)

	if seaArea == "" {
		fmt.Println("Помилка: назва акваторії не може бути порожньою.")
		return
	}

	values, err := waves.Generate(count, seaArea)

	if err != nil {
		fmt.Println("Помилка генерації:", err)
		return
	}

	fmt.Printf("\nАкваторія: %s\n", seaArea)
	fmt.Println("Висота хвиль, м:", values)

	fmt.Print("Введіть значення висоти хвилі: ")

	var limit float64
	if _, err := fmt.Fscan(in, &limit); err != nil {
		fmt.Println("Помилка введення висоти:", err)
		return
	}

	first := 0

	for i, h := range values {
		if h > limit {
			first = i + 1
			break
		}
	}

	if first == 0 {
		fmt.Println("Таких хвиль немає")
	} else {
		fmt.Println("Номер першої хвилі:", first)
	}

	minV, maxV := values[0], values[0]
	sum := 0.0

	counts := map[string]int{
		"спокійно":       0,
		"невеликі хвилі": 0,
		"значні":         0,
		"небезпечні":     0,
	}

	for _, v := range values {
		if v < minV {
			minV = v
		}

		if v > maxV {
			maxV = v
		}

		sum += v
		counts[category(v)]++
	}

	avg := sum / float64(len(values))

	fmt.Printf("Мінімум: %.3f м\n", minV)
	fmt.Printf("Максимум: %.3f м\n", maxV)
	fmt.Printf("Сума: %.3f м\n", sum)
	fmt.Printf("Середнє: %.3f м\n", avg)

	fmt.Println("\nКатегорії:")

	for _, name := range []string{
		"спокійно",
		"невеликі хвилі",
		"значні",
		"небезпечні",
	} {
		fmt.Printf("%s: %d\n", name, counts[name])
	}

	pairCount := 0

	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] == values[j] {
				pairCount++
			}
		}
	}

	fmt.Println("\nПовтори (вкладені цикли):", pairCount)

	freq := make(map[float64]int)

	for _, v := range values {
		freq[v]++
	}

	keys := make([]float64, 0, len(freq))

	for v := range freq {
		keys = append(keys, v)
	}

	sort.Float64s(keys)

	fmt.Println("Повтори (map):")

	mapPairs := 0

	for _, v := range keys {
		if freq[v] > 1 {
			fmt.Printf("%.3f м — %d раз(и)\n", v, freq[v])
			mapPairs += freq[v] * (freq[v] - 1) / 2
		}
	}

	fmt.Println("Кількість однакових пар (map):", mapPairs)

	seed := int64(42)

	a, err1 := waves.GenerateWithSeed(count, seaArea, seed)
	b, err2 := waves.GenerateWithSeed(count, seaArea, seed)

	if err1 != nil || err2 != nil {
		fmt.Println("Помилка перевірки seed:", err1, err2)
	} else {
		fmt.Println("\nПеревірка однакового seed:",
			reflect.DeepEqual(a, b))
	}

	if counts["небезпечні"] > 0 {
		fmt.Println("\n!!! ПОПЕРЕДЖЕННЯ !!!")
		fmt.Println("Виявлено небезпечну хвилю понад 5 м.")
	}
}
