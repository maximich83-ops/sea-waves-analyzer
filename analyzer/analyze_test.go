package analyzer

import (
	"reflect"
	"testing"
)

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
			name:    "lab2 manual example",
			heights: []float64{0.4, 1.2, 3.0, 6.0, 1.2},
			// сума 11,8; середнє 11,8/5 = 2,36
			wantStats:  Statistics{Count: 5, Sum: 11.8, Min: 0.4, Max: 6.0, Mean: 2.36},
			wantCats:   [4]int{1, 2, 1, 1},
			wantRep:    []Repeat{{Height: 1.2, Occurrences: 2}},
			wantPairs:  1,
			wantDanger: true,
		},
		{
			name:    "category boundaries",
			heights: []float64{0.0, 0.5, 0.6, 2.0, 2.1, 5.0, 5.1, 12.0},
			// сума 27,3; середнє 27,3/8 = 3,4125
			wantStats:  Statistics{Count: 8, Sum: 27.3, Min: 0.0, Max: 12.0, Mean: 3.4125},
			wantCats:   [4]int{2, 2, 2, 2},
			wantRep:    []Repeat{},
			wantPairs:  0,
			wantDanger: true,
		},
		{
			name:    "no repeats and no dangerous waves",
			heights: []float64{1.0, 2.5, 3.5, 4.5, 0.3},
			// сума 11,8; середнє 2,36
			wantStats:  Statistics{Count: 5, Sum: 11.8, Min: 0.3, Max: 4.5, Mean: 2.36},
			wantCats:   [4]int{1, 1, 3, 0},
			wantRep:    []Repeat{},
			wantPairs:  0,
			wantDanger: false,
		},
		{
			name:    "several repeats",
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
				t.Fatalf("NewWaveSeries: unexpected error: %v", err)
			}
			r := s.Analyze()

			st := r.Stats
			if st.Count != tt.wantStats.Count || !almostEqual(st.Sum, tt.wantStats.Sum) ||
				!almostEqual(st.Min, tt.wantStats.Min) || !almostEqual(st.Max, tt.wantStats.Max) ||
				!almostEqual(st.Mean, tt.wantStats.Mean) {
				t.Errorf("Stats = %+v, want %+v", st, tt.wantStats)
			}

			wantOrder := AllCategories()
			for i, cc := range r.Categories {
				if cc.Category != wantOrder[i] || cc.Count != tt.wantCats[i] {
					t.Errorf("Categories[%d] = %v:%d, want %v:%d",
						i, cc.Category, cc.Count, wantOrder[i], tt.wantCats[i])
				}
			}

			if !reflect.DeepEqual(r.Repeats, tt.wantRep) {
				t.Errorf("Repeats = %+v, want %+v", r.Repeats, tt.wantRep)
			}
			if got := r.TotalPairs(); got != tt.wantPairs {
				t.Errorf("TotalPairs() = %d, want %d", got, tt.wantPairs)
			}
			if got := r.HasDangerous(); got != tt.wantDanger {
				t.Errorf("HasDangerous() = %v, want %v", got, tt.wantDanger)
			}
		})
	}
}
