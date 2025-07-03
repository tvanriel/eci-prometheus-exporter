// Package main implements the ECI Prometheus Exporter.
// 
// Copyright 2025 Ted van Riel
//
// Licensed under the EUPL, Version 1.2 
//
// You may not use this work except in compliance with the Licence.
//
// You may obtain a copy of the Licence at:
//
//    https://joinup.ec.europa.eu/software/page/eupl
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the Licence is distributed on an "AS IS" basis, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// Licence for the specific language governing permissions and limitations
// under the Licence.
// SPDX-License-Identifier: EUPL-1.2
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// APIResponse defines the expected structure of the ECI API response.
type APIResponse struct {
	SignatureCount int `json:"signatureCount"`
	Goal           int `json:"goal"`
}

// Application contains the application logic.
type Application struct {
	Initiatives []string
	APIURL      string

	Logger *zap.Logger

	Address    string
	Interval   time.Duration
	HTTPClient *http.Client

	SignatureCount *prometheus.GaugeVec
	SignatureGoal  *prometheus.GaugeVec
	APIDurationVec *prometheus.HistogramVec
}

// NewApplication constructs an application from the configuration.
func NewApplication(
	logger *zap.Logger,
	apiURL string,
	initiatives []string,
	address string,
	httpClient *http.Client,
) *Application {
	var (
		signatureCountVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "eci_signatures",
			Help: "Total number of signatures collected by the European Citizens Initiative",
		}, []string{"initiative_id"})

		signatureGoalVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "eci_signature_goal",
			Help: "Target number of signatures for the European Citizens Initiative",
		}, []string{"initiative_id"})

		apiDurationVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "eci_api_duration_seconds",
			Help:    "Duration of API calls to the ECI endpoint per initiative",
			Buckets: prometheus.DefBuckets,
		}, []string{"initiative_id"})
	)

	return &Application{
		Initiatives: initiatives,
		APIURL:      apiURL,

		Logger:     logger,
		HTTPClient: httpClient,
		Address:    address,

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
func (a *Application) FetchAndUpdateMetrics(ctx context.Context, initiativeID string) error {
	apiURL := fmt.Sprintf("%s%s/public/api/report/progression", a.APIURL, initiativeID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("make request: %w", err)
	}

	logger := a.Logger.With(zap.String("initiative_id", initiativeID))

	timer := prometheus.NewTimer(a.APIDurationVec.WithLabelValues(initiativeID))

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

	var data APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.Error("Failed to decode JSON", zap.Error(err))

		return fmt.Errorf("decode json: %w", err)
	}

	logger.Info("Fetched ECI stats",
		zap.Int("signature_count", data.SignatureCount),
		zap.Int("goal", data.Goal),
		zap.Duration("duration", duration),
	)

	a.SignatureCount.WithLabelValues(initiativeID).Set(float64(data.SignatureCount))
	a.SignatureGoal.WithLabelValues(initiativeID).Set(float64(data.Goal))

	return nil
}

// Serve starts the HTTP server.
func (a *Application) Serve() error {
	sm := http.NewServeMux()
	sm.Handle("/metrics", promhttp.Handler())

	a.Logger.Info("Serving Prometheus metrics", zap.String("endpoint", "/metrics"))

	server := &http.Server{
		ReadTimeout: defaultReadTimeout,
		Handler:     sm,
	}

	l, err := net.Listen("tcp", a.Address)
	if err != nil {
		a.Logger.Error("Cannot listen on configured port", zap.String("address", a.Address))

		return fmt.Errorf("start application listener: %w", err)
	}

	err = server.Serve(l)
	if err != nil {
		a.Logger.Error("HTTP server failed", zap.Error(err))

		return fmt.Errorf("serve application server: %w", err)
	}

	return nil
}

func (a *Application) startPolling(initiativeID string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), interval)
	// Initial fetch
	_ = a.FetchAndUpdateMetrics(ctx, initiativeID)

	cancel()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		_ = a.FetchAndUpdateMetrics(ctx, initiativeID)

		cancel()
	}
}

const (
	defaultInterval    = 5 * time.Minute
	defaultReadTimeout = 3 * time.Second
)

func main() {
	initiativeList := flag.String("initiatives", "", "Comma-separated list of initiative IDs (e.g. 043,045,098)")
	address := flag.String("listen-address", ":8080", "Address to expose Prometheus metrics")
	interval := flag.Duration("interval", defaultInterval, "Polling interval for API updates")
	apiURL := flag.String("api-url", "https://eci.ec.europa.eu/", "The URL to the ECI API")
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync() //nolint:errcheck // don't care.

	initiatives := strings.Split(*initiativeList, ",")
	if len(initiatives) == 0 || (len(initiatives) == 1 && initiatives[0] == "") {
		logger.Fatal("No initiative IDs provided. Use -initiatives flag (e.g. -initiatives=045,098)")
	}

	logger.Info("Starting ECI Exporter",
		zap.Strings("initiatives", initiatives),
		zap.String("listen_address", *address),
		zap.Duration("interval", *interval),
	)

	a := NewApplication(
		logger,
		*apiURL,
		initiatives,
		*address,
		http.DefaultClient,
	)

	// Start one goroutine per initiative
	for _, id := range initiatives {
		id := strings.TrimSpace(id)
		if id == "" {
			continue
		}

		go a.startPolling(id, *interval)
	}

	a.MustRegisterWith(prometheus.DefaultRegisterer)

	err = a.Serve()
	if err != nil {
		logger.Fatal("Run server", zap.Error(err))
	}
}
