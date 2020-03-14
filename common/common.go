package common

import (
	"fmt"
	"math"
	"net/url"
)

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

func UrlJoin(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(uri).String()
}

func PrintColorOutputs(input []rune) string {
	var result string
	for i := 0; i < len(input); i++ {
		r, g, b := rgb(i)
		result += fmt.Sprintf("\033[38;2;%d;%d;%dm%c\033[0m", r, g, b, input[i])
	}
	return result
}

func rgb(i int) (int, int, int) {
	f := 0.1
	return int(math.Sin(f*float64(i)+0)*127 + 128),
		int(math.Sin(f*float64(i)+2*math.Pi/3)*127 + 128),
		int(math.Sin(f*float64(i)+4*math.Pi/3)*127 + 128)
}
