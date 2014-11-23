go-jaro-winkler-distance [![GoDoc](https://godoc.org/github.com/9uuso/go-jaro-winkler-distance?status.png)](https://godoc.org/github.com/9uuso/go-jaro-winkler-distance)
=====

Native [Jaro-Winkler distance](https://en.wikipedia.org/wiki/Jaro%E2%80%93Winkler_distance) in Go. Makes heavy use of strings package, but single query doesn't take longer than about 30us.

Jaro-Winkler distance calculates the "closeness" of two strings and returns a score ranging between 0 to 1.

For example comparing words `DIXON` and `DICKSONX` gives you a score of `0.8133333333333332`.

### Example

	package main

	import (
		"fmt"

		"github.com/9uuso/go-jaro-winkler-distance"
	)

	func main() {
		// See more example strings at http://www.amstat.org/sections/srms/Proceedings/papers/1990_056.pdf
		fmt.Println(jwd.Calculate("DIXON", "DICKSONX"))
	}