package analyzer

import (
	"math"
	"strings"
	"testing"
)

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
			t.Errorf("ValidateCount(%d) = %v, want ok=%v", tt.count, err, tt.wantOK)
		}
	}
}

// TestNormalizeSeaArea — очищення пробілів і межа 100 символів (у рунах, не байтах).
func TestNormalizeSeaArea(t *testing.T) {
	got, err := NormalizeSeaArea("  Азовське море  ")
	if err != nil || got != "Азовське море" {
		t.Errorf("NormalizeSeaArea = %q, %v; want %q, nil", got, err, "Азовське море")
	}
	if _, err := NormalizeSeaArea(strings.Repeat("я", 100)); err != nil {
		t.Errorf("100 Cyrillic characters must be accepted, got %v", err)
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
			t.Errorf("ValidateLimit(%v) = %v, want ok=%v", tt.limit, err, tt.wantOK)
		}
	}
}
