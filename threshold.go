package main

import (
	"time"
)

type (
	// MemberCountryCode is a two-letter string that represents the name of a member country.
	MemberCountryCode string
	// Threshold is a map of thresholds per country.
	Threshold map[MemberCountryCode]int
)

// GetThresholds gives you the per-country goals for Initiatives based on their registration date.
// This was extracted from the Javascript code on the ECI web portal.
//
//nolint:mnd,funlen // Intentional
func GetThresholds(registrationDate time.Time) Threshold {
	switch {
	case registrationDate.After(time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)):
		return Threshold{
			"at": 14400,
			"be": 15840,
			"bg": 12240,
			"cy": 4320,
			"cz": 15120,
			"dk": 10800,
			"ee": 5040,
			"fi": 10800,
			"fr": 58320,
			"de": 69120,
			"gr": 15120,
			"hu": 15120,
			"ie": 10080,
			"it": 54720,
			"lv": 6480,
			"lt": 7920,
			"lu": 4320,
			"mt": 4320,
			"nl": 22320,
			"pl": 38160,
			"pt": 15120,
			"ro": 23760,
			"sk": 10800,
			"si": 6480,
			"es": 43920,
			"se": 15120,
			"hr": 8640,
		}
	case registrationDate.After(time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)):
		return Threshold{
			"at": 13395,
			"be": 14805,
			"bg": 11985,
			"cy": 4230,
			"cz": 14805,
			"dk": 9870,
			"ee": 4935,
			"fi": 9870,
			"fr": 55695,
			"de": 67680,
			"gr": 14805,
			"hu": 14805,
			"ie": 9165,
			"it": 53580,
			"lv": 5640,
			"lt": 7755,
			"lu": 4230,
			"mt": 4230,
			"nl": 20445,
			"pl": 36660,
			"pt": 14805,
			"ro": 23265,
			"sk": 9870,
			"si": 5640,
			"es": 41595,
			"se": 14805,
			"hr": 8460,
		}
	case registrationDate.After(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)):
		return Threshold{
			"at": 13518,
			"be": 15771,
			"bg": 12767,
			"cy": 4506,
			"cz": 15771,
			"dk": 9763,
			"ee": 4506,
			"fi": 9763,
			"fr": 55574,
			"de": 72096,
			"gr": 15771,
			"hu": 15771,
			"ie": 8261,
			"it": 54823,
			"lv": 6008,
			"lt": 8261,
			"lu": 4506,
			"mt": 4506,
			"nl": 19526,
			"pl": 38301,
			"pt": 15771,
			"ro": 24032,
			"sk": 9763,
			"si": 6008,
			"es": 40554,
			"se": 15020,
			"gb": 54823,
			"hr": 8261,
		}
	case registrationDate.After(time.Date(2014, 7, 1, 0, 0, 0, 0, time.UTC)):
		return Threshold{
			"at": 13500,
			"be": 15750,
			"bg": 12750,
			"cy": 4500,
			"cz": 15750,
			"dk": 9750,
			"ee": 4500,
			"fi": 9750,
			"fr": 55500,
			"de": 72000,
			"gr": 15750,
			"hu": 15750,
			"ie": 8250,
			"it": 54750,
			"lv": 6000,
			"lt": 8250,
			"lu": 4500,
			"mt": 4500,
			"nl": 19500,
			"pl": 38250,
			"pt": 15750,
			"ro": 24000,
			"sk": 9750,
			"si": 6000,
			"es": 40500,
			"se": 15000,
			"gb": 54750,
			"hr": 8250,
		}
	case registrationDate.After(time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC)):
		return Threshold{
			"at": 14250,
			"be": 16500,
			"bg": 13500,
			"cy": 4500,
			"cz": 16500,
			"dk": 9750,
			"ee": 4500,
			"fi": 9750,
			"fr": 55500,
			"de": 74250,
			"gr": 16500,
			"hu": 16500,
			"ie": 9000,
			"it": 54750,
			"lv": 6750,
			"lt": 9000,
			"lu": 4500,
			"mt": 4500,
			"nl": 19500,
			"pl": 38250,
			"pt": 16500,
			"ro": 24750,
			"sk": 9750,
			"si": 6000,
			"es": 40500,
			"se": 15000,
			"gb": 54750,
			"hr": 9000,
		}
	default:
		return nil
	}
}
