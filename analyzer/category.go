package analyzer

import "fmt"

// WaveCategory — предметний тип категорії висоти хвилі.
// Порядок констант задає порядок виведення категорій.
type WaveCategory int

const (
	Calm        WaveCategory = iota // спокійно: до 0,5 м
	Small                           // невеликі хвилі: 0,6–2 м
	Significant                     // значні: 2,1–5 м
	Dangerous                       // небезпечні: понад 5 м
)

// AllCategories повертає категорії в порядку, визначеному умовою.
func AllCategories() []WaveCategory {
	return []WaveCategory{Calm, Small, Significant, Dangerous}
}

// String повертає назву категорії українською.
func (c WaveCategory) String() string {
	switch c {
	case Calm:
		return "спокійно"
	case Small:
		return "невеликі хвилі"
	case Significant:
		return "значні"
	case Dangerous:
		return "небезпечні"
	default:
		return fmt.Sprintf("невідома категорія (%d)", int(c))
	}
}

// Classify визначає категорію висоти. Межі включні: 0,5 → спокійно,
// 2,0 → невеликі хвилі, 5,0 → значні, усе більше 5,0 → небезпечні.
// Значення між межами таблиці (наприклад 0,55) потрапляють у вищу категорію.
func Classify(height float64) WaveCategory {
	switch {
	case height <= 0.5:
		return Calm
	case height <= 2.0:
		return Small
	case height <= 5.0:
		return Significant
	default:
		return Dangerous
	}
}

// CategoryCount — кількість хвиль однієї категорії.
type CategoryCount struct {
	Category WaveCategory
	Count    int
}

// Categories повертає кількість хвиль кожної категорії у сталому порядку
// AllCategories(), включно з категоріями, де кількість дорівнює нулю.
func (s WaveSeries) Categories() []CategoryCount {
	counts := make(map[WaveCategory]int, 4)
	for _, h := range s.heights {
		counts[Classify(h)]++
	}
	out := make([]CategoryCount, 0, 4)
	for _, c := range AllCategories() {
		out = append(out, CategoryCount{Category: c, Count: counts[c]})
	}
	return out
}
