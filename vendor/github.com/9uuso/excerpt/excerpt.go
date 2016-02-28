package excerpt

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/kennygrant/sanitize"
)

// Make generates excerpt with word length w from input s.
func Make(s string, w int) string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanWords)
	count := 0
	var excerpt bytes.Buffer
	for scanner.Scan() && count < w {
		count++
		excerpt.WriteString(scanner.Text() + " ")
	}
	return sanitize.HTML(strings.TrimSpace(excerpt.String()))
}
