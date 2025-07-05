package main_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	eci "github.com/tvanriel/eci-prometheus-exporter"
)

func TestParseRegistrationNumber(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		// Named input parameters for target function.
		rn      string
		want    *eci.RegistrationNumber
		wantErr assert.ErrorAssertionFunc
	}{
		"normal registration number": {
			rn:      "ECI(2024)000007",
			want:    &eci.RegistrationNumber{Auth: "ECI", Year: "2024", Number: "000007"},
			wantErr: assert.NoError,
		},
		"corrupt date": {
			rn:      "ECI(nope)000007",
			want:    nil,
			wantErr: errContains("invalid format"),
		},
		"missing registration number": {
			rn:      "ECI(2024)",
			want:    nil,
			wantErr: errContains("invalid format"),
		},
		"missing date": {
			rn:      "ECI()000007",
			want:    nil,
			wantErr: errContains("invalid format"),
		},
		"missing authority": {
			rn:      "(2024)000007",
			want:    nil,
			wantErr: errContains("invalid format"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := eci.ParseRegistrationNumber(tt.rn)

			tt.wantErr(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func MustParseRegistrationNumber(str string) *eci.RegistrationNumber {
	rn, err := eci.ParseRegistrationNumber(str)
	if err != nil {
		panic(fmt.Errorf("must parse registration number failed: %w", err))
	}

	return rn
}
