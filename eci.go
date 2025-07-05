// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// ProgressResponse is the type of response that is returned from the ECI API.
type ProgressResponse struct {
	RegistrationDate string    `json:"registrationDate"`
	SOSReport        SOSReport `json:"sosReport"`
}

type SOSEntry struct {
	CountryCode string `json:"countryCodeType"` // e.g. "NL"
	Total       int    `json:"total"`
}
type SOSReport struct {
	TotalSignatures int        `json:"totalSignatures"`
	UpdateDate      string     `json:"updateDate"` // e.g. "04/06/2025"
	Entries         []SOSEntry `json:"entry"`
}

// Application contains the application logic.
type Application struct {
	Initiatives []RegistrationNumber
	APIURL      string

	Logger *zap.Logger

	Address    string
	Interval   time.Duration
	HTTPClient *http.Client

	HTTPServer *http.Server

	SignatureCount *prometheus.GaugeVec
	SignatureGoal  *prometheus.GaugeVec
	APIDurationVec *prometheus.HistogramVec
}

// NewApplication constructs an application from the configuration.
func NewApplication(
	logger *zap.Logger,
	apiURL string,
	initiatives []RegistrationNumber,
	address string,
	httpClient *http.Client,
) *Application {
	var (
		signatureCountVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "eci_signatures",
			Help: "Number of signatures collected by the European Citizens Initiative Per Country",
		}, []string{"initiative_id", "country_code"})

		signatureGoalVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "eci_signature_threshold",
			Help: "Threshold number of signatures for the European Citizens Initiative",
		}, []string{"initiative_id", "country_code"})

		apiDurationVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "eci_api_duration_seconds",
			Help:    "Duration of API calls to the ECI endpoint per initiative",
			Buckets: prometheus.DefBuckets,
		}, []string{"initiative_id"})
	)

	sm := http.NewServeMux()
	sm.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		ReadTimeout: defaultReadTimeout,
		Handler:     sm,
	}

	return &Application{
		Initiatives: initiatives,
		APIURL:      apiURL,

		Logger:     logger,
		HTTPClient: httpClient,
		Address:    address,
		HTTPServer: server,

		SignatureCount: signatureCountVec,
		SignatureGoal:  signatureGoalVec,
		APIDurationVec: apiDurationVec,
	}
}

// MustRegisterWith registers the application metrics with the given prometheus registerer.
func (a *Application) MustRegisterWith(r prometheus.Registerer) {
	r.MustRegister(a.APIDurationVec, a.SignatureCount, a.SignatureGoal)
}

// ErrNon200 is returned when a non-200 response was given by the ECI API.
var ErrNon200 = errors.New("Non-200 response")

// FetchAndUpdateMetrics performs the request and puts the result in the counters.
func (a *Application) FetchAndUpdateMetrics(ctx context.Context, registrationNumber RegistrationNumber) error {
	apiURL := fmt.Sprintf("%s/core/api/register/details/%s/%s", a.APIURL, registrationNumber.Year, registrationNumber.Number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}

	logger := a.Logger.With(zap.String("initiative_id", registrationNumber.String()))

	timer := prometheus.NewTimer(a.APIDurationVec.WithLabelValues(registrationNumber.String()))

	resp, err := a.HTTPClient.Do(req)

	duration := timer.ObserveDuration()

	if err != nil {
		logger.Error("Error fetching ECI API", zap.Error(err))

		return fmt.Errorf("err doing request: %w", err)
	}

	defer resp.Body.Close() //nolint:errcheck // don't really care.

	if resp.StatusCode != http.StatusOK {
		logger.Error("Non-200 response", zap.Int("status_code", resp.StatusCode))

		return ErrNon200
	}

	var data ProgressResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		logger.Error("Failed to decode JSON", zap.Error(err))

		return fmt.Errorf("decode json: %w", err)
	}

	logger.Info("Fetched ECI stats",
		zap.Int("signature_count", data.SOSReport.TotalSignatures),
		zap.Duration("duration", duration),
	)

	registrationDate, err := time.Parse("02/01/2006", data.RegistrationDate)
	if err != nil {
		logger.Error("failed to parse registration date.", zap.Error(err))

		return fmt.Errorf("cannot parse registration date: %w", err)
	}

	th := GetThresholds(registrationDate)

	for _, e := range data.SOSReport.Entries {
		a.SignatureCount.WithLabelValues(registrationNumber.String(), e.CountryCode).Set(float64(e.Total))
		a.SignatureGoal.WithLabelValues(registrationNumber.String(), e.CountryCode).Set(float64(th[MemberCountryCode(strings.ToLower(e.CountryCode))]))
	}

	return nil
}

// Serve starts the HTTP server.
func (a *Application) Serve() error {
	a.Logger.Info("Serving Prometheus metrics", zap.String("endpoint", "/metrics"))

	l, err := net.Listen("tcp", a.Address)
	if err != nil {
		a.Logger.Error("Cannot listen on configured port", zap.String("address", a.Address))

		return fmt.Errorf("start application listener: %w", err)
	}

	err = a.HTTPServer.Serve(l)
	if err != nil {
		a.Logger.Error("HTTP server failed", zap.Error(err))

		return fmt.Errorf("serve application server: %w", err)
	}

	return nil
}

// StartPolling polls when the given ticker ticks.
func (a *Application) StartPolling(registrationNumber RegistrationNumber, ticker *time.Ticker, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = a.FetchAndUpdateMetrics(ctx, registrationNumber)

	cancel()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		_ = a.FetchAndUpdateMetrics(ctx, registrationNumber)

		cancel()
	}
}

const (
	defaultInterval    = 5 * time.Minute
	defaultReadTimeout = 3 * time.Second
)
