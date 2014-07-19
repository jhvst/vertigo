package jwd_test

import (
	"fmt"
	"github.com/9uuso/go-jaro-winkler-distance"
)

func Example() {
	res := jwd.Calculate("DIXON", "DICKSONX")
	fmt.Println(res)
}
