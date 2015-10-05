package timezone

import (
	"fmt"
	"testing"
)

var validCodes = []Timezone{
	{"Europe/Helsinki", "FI", "Finland"},
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

func ExampleValidLocation() {
	fmt.Println(ValidLocation("Europe/Helsinki"))
	// Output: true
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

var expectedOffsets = []int{
	10800,
}

func TestOffset(t *testing.T) {
	for i, expectedOffset := range expectedOffsets {
		zone, offset, err := Offset(validLocations[i])
		if zone != "EEST" {
			t.Errorf("zone of %s should have been %s", validLocations[i], zone)
		}
		if expectedOffset != offset {
			t.Errorf("offset of %s should have been %s", validLocations[i], offset)
		}
		if err != nil {
			t.Errorf("running Offset of %s should not return error, but returned %s", validLocations[i], err)
		}
	}
}

func ExampleOffset() {
	zone, offset, _ := Offset("Europe/Helsinki")
	fmt.Println(zone, offset)
	// Output: EEST 10800
}

func TestCountry(t *testing.T) {
	z, _ := Country("Finland")
	if len(z) != 1 {
		t.Errorf("expected length of Finland to be 1, got %d", len(z))
	}
	if z[0].Code != "FI" {
		t.Errorf("country code of Finland should have been FI, got %s", z[0].Code)
	}
	if z[0].Location != "Europe/Helsinki" {
		t.Errorf("location of Finland should have been Europe/Helsinki, got %s", z[0].Location)
	}
}
