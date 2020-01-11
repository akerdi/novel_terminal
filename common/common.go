package common

// Ternary (true, 1, 2) => 1
func Ternary(expr bool, whenTrue, whenFalse interface{}) interface{} {
	if expr == true {
		return whenTrue
	}
	return whenFalse
}

// Min (1, 2) => 1
func Min(a, b int) int {
	i := Ternary(a < b, a, b)
	r, _ := i.(int)
	return r
}

// Max (1, 2) => 2
func Max(a, b int) int {
	i := Ternary(a >= b, a, b)
	r, _ := i.(int)
	return r
}
