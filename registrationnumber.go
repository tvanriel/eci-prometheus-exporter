package main

import (
	"errors"
	"fmt"
	"regexp"
)

// RegistrationNumber is a parsed registration number.
type RegistrationNumber struct {
	Auth, Year, Number string
}

// String implements [fmt.Stringer].
func (r *RegistrationNumber) String() string {
	return fmt.Sprintf("%s(%s)%s", r.Auth, r.Year, r.Number)
}

// ErrInvalidRegistrationNumber is returned when the registration number layout is invalid.
var ErrInvalidRegistrationNumber = errors.New("invalid format")

// ParseRegistrationNumber reads the Registration number from a string.
func ParseRegistrationNumber(rn string) (*RegistrationNumber, error) {
	re := regexp.MustCompile(`^([A-Z]+)\((\d{4})\)(\d{6})$`)

	matches := re.FindStringSubmatch(rn)

	//nolint:mnd // regex is just above.
	if len(matches) != 4 {
		return nil, ErrInvalidRegistrationNumber
	}

	return &RegistrationNumber{
		Auth:   matches[1],
		Year:   matches[2],
		Number: matches[3],
	}, nil
}
