package analyzer

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
