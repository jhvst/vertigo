package jwd

import (
	"math"
	"strings"
)

// According to this tool: http://www.csun.edu/english/edit_distance.php
// parameter order should have no difference in the result. Therefore,
// to avoid panicing later on, we will order the strings according to
// their length.
func order(s1, s2 string) (string, string) {
	if strings.Count(s1, "")-1 <= strings.Count(s2, "")-1 {
		return s1, s2
	}
	return s2, s1
}

// Calculate calculates Jaro-Winkler distance of two strings. The function lowercases and sorts the parameters
// so that that the longest string is evaluated against the shorter one.
func Calculate(s1, s2 string) float64 {

	s1, s2 = order(strings.ToLower(s1), strings.ToLower(s2))

	// This avoids the function to return NaN.
	if strings.Count(s1, "") == 1 || strings.Count(s2, "") == 1 {
		return float64(0)
	}

	// m as `matching characters`
	// t as `transposition`
	// l as `the length of common prefix at the start of the string up to a maximum of 4 characters`.
	// See more: https://en.wikipedia.org/wiki/Jaro%E2%80%93Winkler_distance
	m := 0
	t := 0
	l := 0

	window := math.Floor(float64(math.Max(float64(len(s1)), float64(len(s2)))/2) - 1)

	//debug:
	//fmt.Println("s1:", s1, "s2:", s2)
	//fmt.Println("Match window:", window)
	//fmt.Println("len(s1):", len(s1), "len(s2):", len(s2))

	for i := 0; i < len(s1); i++ {
		// Exact match
		if s1[i] == s2[i] {
			m++
			// Common prefix limitter
			if i == l && i < 4 {
				l++
			}
		} else {
			if strings.Contains(s2, string(s1[i])) {
				// The character is also considered matching if the amount of characters between the occurances in s1 and s2
				// is less than match window
				gap := strings.Index(s2, string(s1[i])) - strings.Index(s1, string(s1[i]))
				if gap <= int(window) {
					m++
					// Check if transposition is in reach of window
					for k := i; k < len(s1); k++ {
						if strings.Index(s2, string(s1[k])) <= i {
							t++
						}
					}
				}
			}
		}
	}

	distance := (float64(m)/float64(len(s1)) + float64(m)/float64(len(s2)) + (float64(m)-math.Floor(float64(t)/float64(2)))/float64(m)) / float64(3)
	jwd := distance + (float64(l) * float64(0.1) * (float64(1) - distance))

	//debug:
	//fmt.Println("- transpositions:", t)
	//fmt.Println("- matches:", m)
	//fmt.Println("- l:", l)
	//fmt.Println(jwd)

	return jwd

}
