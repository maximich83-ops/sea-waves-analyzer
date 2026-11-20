package analyzer

import (
	"fmt"
	"testing"
)

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
		t.Run(fmt.Sprintf("height %.2f", tt.height), func(t *testing.T) {
			if got := Classify(tt.height); got != tt.want {
				t.Errorf("Classify(%v) = %v, want %v", tt.height, got, tt.want)
			}
		})
	}
}

// TestAnalyze — основний табличний тест зі статичними наборами
