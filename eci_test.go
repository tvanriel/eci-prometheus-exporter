// SPDX-License-Identifier: EUPL-1.2

package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	eci "github.com/tvanriel/eci-prometheus-exporter"
	"go.uber.org/zap/zaptest"
)

type Testserver func(t *testing.T) *httptest.Server

func ServerWantsNoRequests(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { t.Fail() }))
}

func ServerReportsCalls(n *int) Testserver {
	return func(t *testing.T) *httptest.Server {
		t.Helper()

		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			(*n)++
			_, _ = w.Write([]byte(defaultResponse))
		}))
	}
}

//nolint:lll // testdata
const defaultResponse = `{"sosReport":{"totalSignatures":1149248,"updateDate":"04/06/2025","entry":[{"countryCodeType":"SK","total":15885},{"countryCodeType":"EE","total":8272},{"countryCodeType":"BE","total":26068},{"countryCodeType":"BG","total":12266},{"countryCodeType":"IT","total":64528},{"countryCodeType":"IE","total":31377},{"countryCodeType":"FR","total":120251},{"countryCodeType":"ES","total":100283},{"countryCodeType":"HU","total":22569},{"countryCodeType":"SE","total":60153},{"countryCodeType":"DE","total":243448},{"countryCodeType":"SI","total":6072},{"countryCodeType":"LU","total":2387},{"countryCodeType":"MT","total":1753},{"countryCodeType":"LV","total":7004},{"countryCodeType":"DK","total":31965},{"countryCodeType":"CY","total":1898},{"countryCodeType":"AT","total":18513},{"countryCodeType":"GR","total":17513},{"countryCodeType":"NL","total":75811},{"countryCodeType":"LT","total":12783},{"countryCodeType":"CZ","total":19789},{"countryCodeType":"HR","total":12554},{"countryCodeType":"RO","total":32197},{"countryCodeType":"PT","total":27391},{"countryCodeType":"PL","total":126130},{"countryCodeType":"FI","total":50388}]},"registrationDate":"19/06/2024"}`

const unparsableDate = `{"sosReport":{"totalSignatures":0,"entry":[]},"registrationDate":"this is not a valid date"}`

func ServerWantsCallForInitiativeID(rn *eci.RegistrationNumber) Testserver {
	return func(t *testing.T) *httptest.Server {
		t.Helper()

		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, fmt.Sprintf("/core/api/register/details/%s/%s", rn.Year, rn.Number))

			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(defaultResponse))
		}))
	}
}

func NotJSON(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`This is not JSON`))
	}))
}

func InvalidDate(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(unparsableDate))
	}))
}

func BrokenAF(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
}

func errIs(orig error) assert.ErrorAssertionFunc {
	return func(tt assert.TestingT, err error, _ ...any) bool {
		return assert.ErrorIs(tt, orig, err)
	}
}

func errContains(text string) assert.ErrorAssertionFunc {
	return func(tt assert.TestingT, err error, _ ...any) bool {
		return assert.Contains(tt, err.Error(), text)
	}
}

//nolint:funlen // long test.
func TestApplication_FetchAndUpdateMetrics(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		initiatives []eci.RegistrationNumber
		address     string
		httpClient  *http.Client

		wantErr assert.ErrorAssertionFunc

		server Testserver
	}{
		"no initiatives to loop over": {
			initiatives: []eci.RegistrationNumber{},
			server:      ServerWantsNoRequests,
			wantErr:     assert.NoError,
		},

		"correct initiative gets request": {
			initiatives: []eci.RegistrationNumber{*MustParseRegistrationNumber("ECI(2024)000007")},
			server:      ServerWantsCallForInitiativeID(MustParseRegistrationNumber("ECI(2024)000007")),
			wantErr:     assert.NoError,
		},

		"non-200 error returns non-200 response": {
			initiatives: []eci.RegistrationNumber{*MustParseRegistrationNumber("ECI(2024)000007")},
			server:      BrokenAF,
			wantErr:     errIs(eci.ErrNon200),
		},
		"non-json errors correctly": {
			initiatives: []eci.RegistrationNumber{*MustParseRegistrationNumber("ECI(2024)000007")},
			server:      NotJSON,
			wantErr:     errContains("decode json"),
		},
		"unparsable registration date": {
			initiatives: []eci.RegistrationNumber{*MustParseRegistrationNumber("ECI(2024)000007")},
			server:      InvalidDate,
			wantErr:     errContains("parse registration date"),
		},
	}
	for name := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := tests[name].server(t)
			defer server.Close()

			logger := zaptest.NewLogger(t)
			a := eci.NewApplication(
				logger,
				server.URL,

				tests[name].initiatives,
				tests[name].address,
				http.DefaultClient,
			)

			for _, i := range tests[name].initiatives {
				gotErr := a.FetchAndUpdateMetrics(t.Context(), i)
				tests[name].wantErr(t, gotErr)
			}
		})
	}
}

func TestApplication_MustRegisterWith(t *testing.T) {
	t.Parallel()

	a := eci.NewApplication(zaptest.NewLogger(t), "", []eci.RegistrationNumber{}, ":8080", http.DefaultClient)

	assert.NotPanics(t, func() {
		a.MustRegisterWith(prometheus.NewRegistry())
	})
}

func TestApplication_StartPolling(t *testing.T) {
	t.Parallel()

	counter := 0
	server := ServerReportsCalls(&counter)(t)

	log := zaptest.NewLogger(t)
	app := eci.NewApplication(
		log,
		server.URL+"/",
		[]eci.RegistrationNumber{},
		":8080",
		http.DefaultClient,
	)

	ticker := time.NewTicker(100 * time.Millisecond)

	go func() {
		app.StartPolling(*MustParseRegistrationNumber("ECI(2024)000007"), ticker, 100*time.Millisecond)
	}()

	assert.EventuallyWithT(t, func(collect *assert.CollectT) {
		assert.GreaterOrEqual(collect, counter, 3, "expected multiple polling calls")
	}, 350*time.Millisecond, 100*time.Millisecond)

	ticker.Stop()
	time.Sleep(100 * time.Millisecond) // flake
}

//nolint:paralleltest // do not run me parallel.
func TestApplication_Serve(t *testing.T) {
	server := ServerWantsCallForInitiativeID(MustParseRegistrationNumber("ECI(2024)000007"))(t)
	app := eci.NewApplication(
		zaptest.NewLogger(t),
		server.URL,
		[]eci.RegistrationNumber{*MustParseRegistrationNumber("ECI(2024)000007")},
		":12415",
		http.DefaultClient,
	)

	go func() {
		_ = app.Serve()
	}()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:12415/metrics", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Body)

	_ = resp.Body.Close()
}

//nolint:paralleltest // do not run me parallel.
func TestApplication_ServeInvalidListenAddre(t *testing.T) {
	server := ServerWantsCallForInitiativeID(MustParseRegistrationNumber("ECI(2024)000007"))(t)
	app := eci.NewApplication(
		zaptest.NewLogger(t),
		server.URL,
		[]eci.RegistrationNumber{*MustParseRegistrationNumber("ECI(2024)000007")},
		"i am not a valid addr.",
		http.DefaultClient,
	)

	require.Error(t, app.Serve())
}
