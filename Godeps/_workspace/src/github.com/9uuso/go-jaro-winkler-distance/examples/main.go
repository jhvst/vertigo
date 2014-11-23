package main

import (
	"fmt"

	"github.com/9uuso/go-jaro-winkler-distance"
)

func main() {
	// See more example strings at http://www.amstat.org/sections/srms/Proceedings/papers/1990_056.pdf
	fmt.Println(jwd.Calculate("DIXON", "DICKSONX"))
}
