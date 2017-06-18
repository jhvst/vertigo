package jwd

import (
	"bytes"
	"math"
	"strings"
	"unicode/utf8"
)

const weight = 0.1
const commonPrefixLimiter = 4

// To avoid panicing later on, order strings according to
// their unicode length.
func sort(s1, s2 string) (shorter, longer string) {
	if utf8.RuneCountInString(s1) < utf8.RuneCountInString(s2) {
		return s1, s2
	}
	return s2, s1
}

func window(s2 string) float64 {
	return math.Floor(float64(utf8.RuneCountInString(s2)/2 - 1))
}

func score(m, t, runes1len, runes2len float64) float64 {
	return (m/runes1len + m/runes2len + (m-math.Floor(t/2))/m) / 3
}

func naiveSearchDescending(s []rune, searched rune, start int) int {
	for i := start; i > 0; i-- {
		if s[i] == searched {
			return i
		}
	}
	return strings.Index(string(s), string(searched))
}

func naiveSearchAscending(s []rune, searched rune, start int) int {
	for i := start; i < len(s); i++ {
		if s[i] == searched {
			return i
		}
	}
	return strings.LastIndex(string(s), string(searched))
}

// closestIndex returns position of the closest rune r starting
// from pos
func closestIndex(s []rune, r rune, pos int) int {

	desc := naiveSearchDescending(s, r, pos)
	asc := naiveSearchAscending(s, r, pos)

	da := math.Abs(float64(desc - pos))
	aa := math.Abs(float64(asc - pos))

	if da < aa {
		return desc
	}
	return asc
}

// Calculate calculates Jaro-Winkler distance of two strings.
// The function lowercases its parameters.
func Calculate(s1, s2 string) float64 {

	// Avoid returning NaN
	if utf8.RuneCountInString(s1) == 0 || utf8.RuneCountInString(s2) == 0 {
		return 0
	}

	s1, s2 = sort(strings.ToLower(s1), strings.ToLower(s2))

	// m as `matching characters`
	// t as `transposition`
	// l as `the length of common prefix at the start of the string up to a maximum of 4 characters`.
	// See more: https://en.wikipedia.org/wiki/Jaro%E2%80%93Winkler_distance
	var m, t, l float64

	window := window(s2)

	runes1 := bytes.Runes([]byte(s1))
	runes2 := bytes.Runes([]byte(s2))

	for i := 0; i < len(runes1); i++ {
		// Exact match
		if runes1[i] == runes2[i] {
			m++
			// Common prefix limiter
			if i == int(l) && i <= commonPrefixLimiter {
				l++
			}
		} else if strings.Contains(s2, string(runes1[i])) {

			// The character is also considered matching if the amount of characters between the occurances in s1 and s2
			// is less than match window
			c := closestIndex(runes2, runes1[i], i)

			gap := math.Abs(float64(c - i))

			//debug:
			//fmt.Println("searched rune", string(runes1[i]), "with starting index", i, "from word", string(runes2), "-- closest index was", c)
			//fmt.Println("string", s2, "contains", string(runes1[i]))

			if gap <= window {
				m++
				// Check if transposition is in reach of window
				for k := i; k < len(runes1); k++ {
					if closestIndex(runes2, runes1[k], i) <= i {
						t++
					}
				}
			}
		}
	}

	score := score(m, t, float64(len(runes1)), float64(len(runes2)))
	distance := score + (l * weight * (1 - score))

	if math.IsNaN(distance) {
		return 0
	}

	//debug:
	//fmt.Println("- score:", score)
	//fmt.Println("- transpositions:", t)
	//fmt.Println("- matches:", m)
	//fmt.Println("- window:", window)
	//fmt.Println("- l:", l)
	//fmt.Println("- distance:", distance)

	return distance
}
