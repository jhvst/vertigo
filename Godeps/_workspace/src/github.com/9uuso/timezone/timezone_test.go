package timezone

import "testing"

var validCodes = []Timezone{
	{"Europe/Helsinki", "FI"},
}

func TestCode(t *testing.T) {

	for _, timezone := range validCodes {

		z, err := Code(timezone.Code)
		if err != nil {
			t.Errorf("%s should have been found, but was not", timezone.Code)
		}
		if len(z) != 1 {
			t.Errorf("%s len should been %d, but it was d", timezone.Location, 1, len(z))
		}
		if z[0].Location != timezone.Location {
			t.Errorf("Location for %s should have been %s, but it was %s", timezone.Code, timezone.Location, z[0].Location)
		}

	}

}

var validLocations = []string{
	"Europe/Helsinki",
}

func TestValidLocation(t *testing.T) {
	for _, loc := range validLocations {
		if !ValidLocation(loc) {
			t.Errorf("%s should have been valid location", loc)
		}
	}
}

var invalidLocations = []string{
	"Europe/Oulu",
	"FI",
}

func TestInvalidLocation(t *testing.T) {
	for _, loc := range invalidLocations {
		if ValidLocation(loc) {
			t.Errorf("%s should not have been valid location", loc)
		}
	}
}
