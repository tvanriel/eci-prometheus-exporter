# European Citizens Initiative Prometheus Exporter

[![Docker Pulls](https://badgen.net/docker/pulls/mitaka8/eci-prometheus-exporter?icon=docker&label=pulls)](https://hub.docker.com/r/mitaka8/eci-prometheus-exporter/)
[![Docker Stars](https://badgen.net/docker/stars/mitaka8/eci-prometheus-exporter?icon=docker&label=stars)](https://hub.docker.com/r/mitaka8/eci-prometheus-exporter/)
[![Docker Image Size](https://badgen.net/docker/size/mitaka8/eci-prometheus-exporter?icon=docker&label=image%20size)](https://hub.docker.com/r/mitaka8/eci-prometheus-exporter/)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/tvanriel/eci-prometheus-exporter)

A tiny, Go service that scrapes the **European Citizens Initiative** statistics API and exposes them.

|  Metric              |  Type    |  Description                                                  |
| -------------------- | -------- | ------------------------------------------------------------- |
| `eci_signatures`     |  `gauge` |  Number of signatures collected by the European Citizens Initiative Per member state.                       |
| `eci_signature_threshold` |  `gauge` |  Threshold number of signatures per member state.     |

---

## Quick Start

### Binary

```bash
go install github.com/tvanriel/eci-prometheus-exporter@latest
eci-prometheus-exporter \
  -initiatives=ECI(2024)000007 \
  -listen-address=:8080 \
  -interval=5m
```

### Docker

```bash
docker run -d --name eci-exporter \
  -p 8080:8080 \
  docker.io/mitaka8/eci-prometheus-exporter:latest \
  -initiatives=ECI(2024)000007
```

### Kubernetes

```bash
kubectl apply -k ./deploy
```
---

## Configuration Flags

| Flag              | Default       | Description                    |
| ----------------- | ------------- | ------------------------------ |
| `-initiatives`     | **required** | Initiative IDs, e.g. `ECI(2024)000007,ECI(2024)000008` |
| `-listen-address` | `:8080`       | HTTP bind address              |
| `-interval`       | `5m`          | Polling interval               |

---

## Development

```bash
go run . -initiatives=ECI(2024)000007
```
---
## Copyright

Copyright 2025 Ted van Riel

Licensed under the EUPL, Version 1.2 

You may not use this work except in compliance with the Licence.

You may obtain a copy of the Licence at:

   https://joinup.ec.europa.eu/software/page/eupl

Unless required by applicable law or agreed to in writing, software
distributed under the Licence is distributed on an "AS IS" basis, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
Licence for the specific language governing permissions and limitations
under the Licence.
