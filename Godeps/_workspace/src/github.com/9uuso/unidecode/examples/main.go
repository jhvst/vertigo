package main

import (
	"fmt"

	"unidecode"
)

func main() {
	fmt.Println(unidecode.Unidecode("áéíóú"))
}
