package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	eci "github.com/tvanriel/eci-prometheus-exporter"
	"go.uber.org/zap/zaptest"
)

type Testserver func(t *testing.T) *httptest.Server

func ServerWantsNoRequests(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { t.Fail() }))
}

const defaultResponse = `{"signatureCount":1030411,"goal":1000000}`

func ServerWantsCallForInitiativeID(id string) Testserver {
	return func(t *testing.T) *httptest.Server {
		t.Helper()

		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, fmt.Sprintf("/%s/public/api/report/progression", id))

			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(defaultResponse))
		}))
	}
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

func TestApplication_FetchAndUpdateMetrics(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		initiatives []string
		address     string
		httpClient  *http.Client

		wantErr assert.ErrorAssertionFunc

		server Testserver
	}{
		"no initiatives to loop over": {
			initiatives: []string{},
			server:      ServerWantsNoRequests,
			wantErr:     assert.NoError,
		},

		"correct initiative gets request": {
			initiatives: []string{"045"},
			server:      ServerWantsCallForInitiativeID("045"),
			wantErr:     assert.NoError,
		},

		"non-200 error returns non-200 response": {
			initiatives: []string{"045"},
			server:      BrokenAF,
			wantErr:     errIs(eci.ErrNon200),
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
				server.URL+"/",

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
