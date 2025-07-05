// Package main implements the ECI Prometheus Exporter.
//
// SPDX-License-Identifier: EUPL-1.2
//
// # Copyright 2025 Ted van Riel
//
// # Licensed under the EUPL, Version 1.2
//
// You may not use this work except in compliance with the Licence.
//
// You may obtain a copy of the Licence at:
//
//	https://joinup.ec.europa.eu/software/page/eupl
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the Licence is distributed on an "AS IS" basis, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// Licence for the specific language governing permissions and limitations
// under the Licence.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	initiativeList := flag.String("initiatives", "", "Comma-separated list of initiative IDs (e.g. 043,045,098)")
	address := flag.String("listen-address", ":8080", "Address to expose Prometheus metrics")
	interval := flag.Duration("interval", defaultInterval, "Polling interval for API updates")
	apiURL := flag.String("api-url", "https://register.eci.ec.europa.eu", "The URL to the ECI API")
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

	registrationNumbers := make([]RegistrationNumber, 0, len(initiatives))
	for _, i := range initiatives {
		rn, err := ParseRegistrationNumber(i)
		if err != nil {
			logger.Fatal("Cannot parse registratin number", zap.String("registration_number", i))
		}

		registrationNumbers = append(registrationNumbers, *rn)
	}

	a := NewApplication(
		logger,
		*apiURL,
		registrationNumbers,
		*address,
		http.DefaultClient,
	)

	// Start one goroutine per initiative
	for _, id := range registrationNumbers {
		ticker := time.NewTicker(*interval)

		go a.StartPolling(id, ticker, *interval)
	}

	a.MustRegisterWith(prometheus.DefaultRegisterer)

	err = a.Serve()
	if err != nil {
		logger.Fatal("Run server", zap.Error(err))
	}
}
