// Package analyzer містить предметну логіку аналізу висот морських хвиль
// (варіант 5 «Морські хвилі»): перевірку вхідних параметрів, категорії,
// статистику та пошук повторів. Пакет не читає аргументи командного рядка,
// нічого не друкує й не завершує процес.
package analyzer

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Межі предметної області варіанта 5.
const (
	MinCount        = 5    // мінімальна кількість значень
	MaxCount        = 100  // максимальна кількість значень
	MinHeight       = 0.0  // мінімальна висота хвилі, м
	MaxHeight       = 12.0 // максимальна висота хвилі, м
	MaxSeaAreaRunes = 100  // максимальна довжина назви акваторії, символів
)

// Помилки перевірки. Їх можна розпізнати через errors.Is.
var (
	ErrCountOutOfRange  = errors.New("кількість значень поза межами")
	ErrEmptySeaArea     = errors.New("назва акваторії порожня")
	ErrSeaAreaTooLong   = errors.New("назва акваторії задовга")
	ErrSeaAreaControl   = errors.New("назва акваторії містить керівні символи")
	ErrHeightOutOfRange = errors.New("висота хвилі поза діапазоном")
	ErrLimitOutOfRange  = errors.New("порогова висота поза діапазоном")
)

// ValidateCount перевіряє кількість значень до виклику генератора.
func ValidateCount(count int) error {
	if count < MinCount || count > MaxCount {
		return fmt.Errorf("%w: %d (дозволено від %d до %d)", ErrCountOutOfRange, count, MinCount, MaxCount)
	}
	return nil
}

// NormalizeSeaArea очищає назву акваторії через strings.TrimSpace і перевіряє її.
func NormalizeSeaArea(raw string) (string, error) {
	area := strings.TrimSpace(raw)
	if area == "" {
		return "", ErrEmptySeaArea
	}
	if n := utf8.RuneCountInString(area); n > MaxSeaAreaRunes {
		return "", fmt.Errorf("%w: %d символів (максимум %d)", ErrSeaAreaTooLong, n, MaxSeaAreaRunes)
	}
	for _, r := range area {
		if unicode.IsControl(r) {
			return "", ErrSeaAreaControl
		}
	}
	return area, nil
}

// ValidateLimit перевіряє порогову висоту для спеціальної перевірки варіанта 5.
func ValidateLimit(limit float64) error {
	if math.IsNaN(limit) || limit < MinHeight || limit > MaxHeight {
		return fmt.Errorf("%w: %v (дозволено %.0f–%.0f м)", ErrLimitOutOfRange, limit, MinHeight, MaxHeight)
	}
	return nil
}

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

// WaveSeries — перевірений набір висот хвиль для однієї акваторії.
// Нульове значення не є корисним: створюйте серію через NewWaveSeries.
type WaveSeries struct {
	seaArea string
	heights []float64
}

// NewWaveSeries створює серію та перевіряє інваріанти:
// назва акваторії непорожня, кількість значень 5–100,
// кожна висота скінченна й лежить у діапазоні 0–12 м.
// Зріз копіюється, тож подальша зміна вхідного зрізу не порушує серію.
func NewWaveSeries(seaArea string, heights []float64) (WaveSeries, error) {
	area, err := NormalizeSeaArea(seaArea)
	if err != nil {
		return WaveSeries{}, err
	}
	if err := ValidateCount(len(heights)); err != nil {
		return WaveSeries{}, err
	}
	for i, h := range heights {
		if math.IsNaN(h) || h < MinHeight || h > MaxHeight {
			return WaveSeries{}, fmt.Errorf("%w: елемент %d = %v (дозволено %.0f–%.0f м)",
				ErrHeightOutOfRange, i, h, MinHeight, MaxHeight)
		}
	}
	copied := make([]float64, len(heights))
	copy(copied, heights)
	return WaveSeries{seaArea: area, heights: copied}, nil
}

// SeaArea повертає назву акваторії.
func (s WaveSeries) SeaArea() string { return s.seaArea }

// Len повертає кількість значень.
func (s WaveSeries) Len() int { return len(s.heights) }

// Heights повертає копію висот, щоб зовнішній код не змінив стан серії.
func (s WaveSeries) Heights() []float64 {
	out := make([]float64, len(s.heights))
	copy(out, s.heights)
	return out
}

// FirstAbove — спеціальна перевірка варіанта 5: повертає номер (з 1)
// першої хвилі, вищої за limit, і true; якщо такої немає — 0 і false.
func (s WaveSeries) FirstAbove(limit float64) (int, bool) {
	for i, h := range s.heights {
		if h > limit {
			return i + 1, true
		}
	}
	return 0, false
}

// Statistics — похідні статистичні показники серії, м.
type Statistics struct {
	Count int
	Sum   float64
	Min   float64
	Max   float64
	Mean  float64
}

// Stats обчислює кількість, суму, мінімум, максимум і середнє без округлення.
func (s WaveSeries) Stats() Statistics {
	st := Statistics{Count: len(s.heights)}
	if st.Count == 0 {
		return st
	}
	st.Min, st.Max = s.heights[0], s.heights[0]
	for _, h := range s.heights {
		if h < st.Min {
			st.Min = h
		}
		if h > st.Max {
			st.Max = h
		}
		st.Sum += h
	}
	st.Mean = st.Sum / float64(st.Count)
	return st
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

// Repeat — значення, що трапляється в серії більше одного разу.
type Repeat struct {
	Height      float64
	Occurrences int
}

// Pairs повертає кількість однакових пар для цього значення: n*(n-1)/2.
func (r Repeat) Pairs() int { return r.Occurrences * (r.Occurrences - 1) / 2 }

// Repeats повертає повторювані значення за зростанням висоти.
func (s WaveSeries) Repeats() []Repeat {
	freq := make(map[float64]int, len(s.heights))
	for _, h := range s.heights {
		freq[h]++
	}
	out := make([]Repeat, 0)
	for h, n := range freq {
		if n > 1 {
			out = append(out, Repeat{Height: h, Occurrences: n})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Height < out[j].Height })
	return out
}

// Report — повний результат аналізу серії.
type Report struct {
	SeaArea    string
	Heights    []float64
	Stats      Statistics
	Categories []CategoryCount
	Repeats    []Repeat
}

// Analyze будує повний звіт для серії.
func (s WaveSeries) Analyze() Report {
	return Report{
		SeaArea:    s.seaArea,
		Heights:    s.Heights(),
		Stats:      s.Stats(),
		Categories: s.Categories(),
		Repeats:    s.Repeats(),
	}
}

// TotalPairs повертає загальну кількість однакових пар у звіті.
func (r Report) TotalPairs() int {
	total := 0
	for _, rep := range r.Repeats {
		total += rep.Pairs()
	}
	return total
}

// HasDangerous повідомляє, чи є хоча б одна небезпечна хвиля (понад 5 м).
func (r Report) HasDangerous() bool {
	for _, cc := range r.Categories {
		if cc.Category == Dangerous {
			return cc.Count > 0
		}
	}
	return false
}
