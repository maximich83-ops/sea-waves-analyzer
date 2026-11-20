package analyzer

import "math"

// eps — допуск для порівняння дробових чисел.
const eps = 1e-9

func almostEqual(a, b float64) bool { return math.Abs(a-b) < eps }
