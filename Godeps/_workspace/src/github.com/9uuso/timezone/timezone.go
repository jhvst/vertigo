package timezone

import "errors"

// ValidLocation returns true if the given string can be found as a
// timezone location in the timezone.Locations array.
func ValidLocation(s string) bool {
	for _, zone := range Locations {
		if zone.Location == s {
			return true
		}
	}
	return false
}

// Code returns all timezones with given country code.
// If none is found, returns an error.
func Code(c string) ([]Timezone, error) {
	var z []Timezone
	for _, zone := range Locations {
		if zone.Code == c {
			z = append(z, zone)
		}
	}
	if len(z) == 0 {
		return z, errors.New("no timezones found")
	}
	return z, nil
}
