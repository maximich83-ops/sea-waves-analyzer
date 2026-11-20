package analyzer

import (
	"errors"
	"math"
	"strings"
	"testing"
)

// TestNewWaveSeriesErrors — негативні випадки конструктора.
func TestNewWaveSeriesErrors(t *testing.T) {
	ok5 := []float64{1, 2, 3, 4, 5}
	tests := []struct {
		name    string
		area    string
		heights []float64
		wantErr error
	}{
		{"empty slice", "Чорне море", []float64{}, ErrCountOutOfRange},
		{"too few values (4)", "Чорне море", []float64{1, 2, 3, 4}, ErrCountOutOfRange},
		{"too many values (101)", "Чорне море", make([]float64, 101), ErrCountOutOfRange},
		{"negative height", "Чорне море", []float64{1, 2, 3, 4, -0.1}, ErrHeightOutOfRange},
		{"height above 12 m", "Чорне море", []float64{1, 2, 3, 4, 12.1}, ErrHeightOutOfRange},
		{"NaN", "Чорне море", []float64{1, 2, 3, 4, math.NaN()}, ErrHeightOutOfRange},
		{"blank sea area", "   ", ok5, ErrEmptySeaArea},
		{"sea area too long", strings.Repeat("я", 101), ok5, ErrSeaAreaTooLong},
		{"control character", "Чорне\x1bморе", ok5, ErrSeaAreaControl},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWaveSeries(tt.area, tt.heights)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
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
		t.Errorf("Heights()[0] = %v after input slice changed, want 1", got)
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
		{"limit below first wave", 0.0, 1, true},
		{"equal value is not above", 1.2, 3, true},
		{"found in the middle", 2.0, 3, true},
		{"last possible", 5.9, 4, true},
		{"none above", 6.0, 0, false},
		{"maximum limit", 12.0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, ok := s.FirstAbove(tt.limit)
			if pos != tt.wantPos || ok != tt.wantOK {
				t.Errorf("FirstAbove(%v) = (%d, %v), want (%d, %v)", tt.limit, pos, ok, tt.wantPos, tt.wantOK)
			}
		})
	}
}
