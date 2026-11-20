// Package analyzer містить предметну логіку аналізу висот морських хвиль
// (варіант 5 «Морські хвилі»): перевірку вхідних параметрів, категорії,
// статистику та пошук повторів. Пакет не читає аргументи командного рядка,
// нічого не друкує й не завершує процес.
package analyzer

import (
	"fmt"
	"math"
)

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
