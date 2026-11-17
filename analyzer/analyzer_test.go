package analyzer

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

const eps = 1e-9

func almostEqual(a, b float64) bool { return math.Abs(a-b) < eps }

// TestClassify перевіряє межі категорій (включні верхні межі).
func TestClassify(t *testing.T) {
	tests := []struct {
		height float64
		want   WaveCategory
	}{
		{0.0, Calm},
		{0.5, Calm},
		{0.55, Small}, // проміжок між 0,5 і 0,6 відносимо до вищої категорії
		{0.6, Small},
		{2.0, Small},
		{2.1, Significant},
		{5.0, Significant},
		{5.1, Dangerous},
		{12.0, Dangerous},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%.2f м → %v", tt.height, tt.want), func(t *testing.T) {
			if got := Classify(tt.height); got != tt.want {
				t.Errorf("Classify(%v) = %v, очікувалось %v", tt.height, got, tt.want)
			}
		})
	}
}

// TestAnalyze — основний табличний тест зі статичними наборами
// та вручну обчисленими очікуваннями.
func TestAnalyze(t *testing.T) {
	tests := []struct {
		name       string
		heights    []float64
		wantStats  Statistics
		wantCats   [4]int // спокійно, невеликі, значні, небезпечні
		wantRep    []Repeat
		wantPairs  int
		wantDanger bool
	}{
		{
			name:    "коректний набір із ЛР2 (ручний приклад)",
			heights: []float64{0.4, 1.2, 3.0, 6.0, 1.2},
			// сума 11,8; середнє 11,8/5 = 2,36
			wantStats:  Statistics{Count: 5, Sum: 11.8, Min: 0.4, Max: 6.0, Mean: 2.36},
			wantCats:   [4]int{1, 2, 1, 1},
			wantRep:    []Repeat{{Height: 1.2, Occurrences: 2}},
			wantPairs:  1,
			wantDanger: true,
		},
		{
			name:    "граничні значення категорій",
			heights: []float64{0.0, 0.5, 0.6, 2.0, 2.1, 5.0, 5.1, 12.0},
			// сума 27,3; середнє 27,3/8 = 3,4125
			wantStats:  Statistics{Count: 8, Sum: 27.3, Min: 0.0, Max: 12.0, Mean: 3.4125},
			wantCats:   [4]int{2, 2, 2, 2},
			wantRep:    []Repeat{},
			wantPairs:  0,
			wantDanger: true,
		},
		{
			name:    "без повторів і без небезпечних хвиль",
			heights: []float64{1.0, 2.5, 3.5, 4.5, 0.3},
			// сума 11,8; середнє 2,36
			wantStats:  Statistics{Count: 5, Sum: 11.8, Min: 0.3, Max: 4.5, Mean: 2.36},
			wantCats:   [4]int{1, 1, 3, 0},
			wantRep:    []Repeat{},
			wantPairs:  0,
			wantDanger: false,
		},
		{
			name:    "кілька повторів",
			heights: []float64{3.0, 1.5, 8.0, 1.5, 3.0, 1.5},
			// сума 18,5; середнє 18,5/6 ≈ 3,0833; 1,5×3 → 3 пари, 3,0×2 → 1 пара
			wantStats:  Statistics{Count: 6, Sum: 18.5, Min: 1.5, Max: 8.0, Mean: 18.5 / 6},
			wantCats:   [4]int{0, 3, 2, 1},
			wantRep:    []Repeat{{Height: 1.5, Occurrences: 3}, {Height: 3.0, Occurrences: 2}},
			wantPairs:  4,
			wantDanger: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewWaveSeries("Чорне море", tt.heights)
			if err != nil {
				t.Fatalf("NewWaveSeries: неочікувана помилка: %v", err)
			}
			r := s.Analyze()

			st := r.Stats
			if st.Count != tt.wantStats.Count || !almostEqual(st.Sum, tt.wantStats.Sum) ||
				!almostEqual(st.Min, tt.wantStats.Min) || !almostEqual(st.Max, tt.wantStats.Max) ||
				!almostEqual(st.Mean, tt.wantStats.Mean) {
				t.Errorf("Stats = %+v, очікувалось %+v", st, tt.wantStats)
			}

			wantOrder := AllCategories()
			for i, cc := range r.Categories {
				if cc.Category != wantOrder[i] || cc.Count != tt.wantCats[i] {
					t.Errorf("Categories[%d] = %v:%d, очікувалось %v:%d",
						i, cc.Category, cc.Count, wantOrder[i], tt.wantCats[i])
				}
			}

			if !reflect.DeepEqual(r.Repeats, tt.wantRep) {
				t.Errorf("Repeats = %+v, очікувалось %+v", r.Repeats, tt.wantRep)
			}
			if got := r.TotalPairs(); got != tt.wantPairs {
				t.Errorf("TotalPairs() = %d, очікувалось %d", got, tt.wantPairs)
			}
			if got := r.HasDangerous(); got != tt.wantDanger {
				t.Errorf("HasDangerous() = %v, очікувалось %v", got, tt.wantDanger)
			}
		})
	}
}

// TestNewWaveSeriesErrors — негативні випадки конструктора.
func TestNewWaveSeriesErrors(t *testing.T) {
	ok5 := []float64{1, 2, 3, 4, 5}
	tests := []struct {
		name    string
		area    string
		heights []float64
		wantErr error
	}{
		{"порожній набір", "Чорне море", []float64{}, ErrCountOutOfRange},
		{"недостатній набір (4)", "Чорне море", []float64{1, 2, 3, 4}, ErrCountOutOfRange},
		{"забагато значень (101)", "Чорне море", make([]float64, 101), ErrCountOutOfRange},
		{"від'ємна висота", "Чорне море", []float64{1, 2, 3, 4, -0.1}, ErrHeightOutOfRange},
		{"висота понад 12 м", "Чорне море", []float64{1, 2, 3, 4, 12.1}, ErrHeightOutOfRange},
		{"NaN", "Чорне море", []float64{1, 2, 3, 4, math.NaN()}, ErrHeightOutOfRange},
		{"порожня акваторія", "   ", ok5, ErrEmptySeaArea},
		{"задовга акваторія", strings.Repeat("я", 101), ok5, ErrSeaAreaTooLong},
		{"керівний символ", "Чорне\x1bморе", ok5, ErrSeaAreaControl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWaveSeries(tt.area, tt.heights)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("помилка = %v, очікувалось %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateCountBounds — межі count: 4 і 101 відхиляються, 5 і 100 приймаються.
func TestValidateCountBounds(t *testing.T) {
	tests := []struct {
		count  int
		wantOK bool
	}{
		{-1, false}, {0, false}, {4, false}, {5, true}, {100, true}, {101, false},
	}
	for _, tt := range tests {
		err := ValidateCount(tt.count)
		if (err == nil) != tt.wantOK {
			t.Errorf("ValidateCount(%d) = %v, очікувалось ok=%v", tt.count, err, tt.wantOK)
		}
	}
}

// TestNormalizeSeaArea — очищення пробілів і межа 100 символів (у рунах, не байтах).
func TestNormalizeSeaArea(t *testing.T) {
	got, err := NormalizeSeaArea("  Азовське море  ")
	if err != nil || got != "Азовське море" {
		t.Errorf("NormalizeSeaArea = %q, %v; очікувалось %q, nil", got, err, "Азовське море")
	}
	if _, err := NormalizeSeaArea(strings.Repeat("я", 100)); err != nil {
		t.Errorf("100 кириличних символів мають прийматися, отримано %v", err)
	}
}

// TestSeriesCopiesInput — серія не залежить від подальших змін вхідного зрізу.
func TestSeriesCopiesInput(t *testing.T) {
	in := []float64{1, 2, 3, 4, 5}
	s, err := NewWaveSeries("Чорне море", in)
	if err != nil {
		t.Fatal(err)
	}
	in[0] = 11
	if got := s.Heights()[0]; got != 1 {
		t.Errorf("Heights()[0] = %v після зміни вхідного зрізу, очікувалось 1", got)
	}
}

// TestFirstAbove — спеціальна перевірка варіанта 5 (номер першої хвилі, вищої за поріг).
func TestFirstAbove(t *testing.T) {
	s, err := NewWaveSeries("Чорне море", []float64{0.4, 1.2, 3.0, 6.0, 1.2})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		limit   float64
		wantPos int
		wantOK  bool
	}{
		{"поріг нижче першої хвилі", 0.0, 1, true},
		{"рівне значення не рахується (строго більше)", 1.2, 3, true},
		{"знайдено в середині", 2.0, 3, true},
		{"остання можлива", 5.9, 4, true},
		{"вищих немає", 6.0, 0, false},
		{"максимальний поріг", 12.0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, ok := s.FirstAbove(tt.limit)
			if pos != tt.wantPos || ok != tt.wantOK {
				t.Errorf("FirstAbove(%v) = (%d, %v), очікувалось (%d, %v)", tt.limit, pos, ok, tt.wantPos, tt.wantOK)
			}
		})
	}
}

// TestValidateLimit — межі порогової висоти.
func TestValidateLimit(t *testing.T) {
	tests := []struct {
		limit  float64
		wantOK bool
	}{
		{-0.1, false}, {0, true}, {12, true}, {12.1, false}, {math.NaN(), false},
	}
	for _, tt := range tests {
		if err := ValidateLimit(tt.limit); (err == nil) != tt.wantOK {
			t.Errorf("ValidateLimit(%v) = %v, очікувалось ok=%v", tt.limit, err, tt.wantOK)
		}
	}
}
