package analyzer

import "sort"

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
