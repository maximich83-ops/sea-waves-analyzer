package analyzer

// Report — повний результат аналізу серії.
type Report struct {
	SeaArea    string
	Heights    []float64
	Stats      Statistics
	Categories []CategoryCount
	Repeats    []Repeat
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
