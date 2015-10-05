package excerpt

import (
	"testing"
)

const ipsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum gravida, ipsum sit amet placerat aliquam, nibh diam lobortis diam, lacinia."

const expectedString = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum gravida, ipsum sit amet placerat aliquam,"

func TestMake(t *testing.T) {
	s := Make(ipsum, 15)
	if s != expectedString {
		t.Errorf("%s should have equal %s, but was %s", ipsum, expectedString, s)
	}
}
