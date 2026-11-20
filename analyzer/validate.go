package analyzer

import (
	"errors"
	"fmt"
	"math"
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
