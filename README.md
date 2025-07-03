# European Citizens Initiative Prometheus Exporter

A tiny, Go service that scrapes the **European Citizens Initiative** statistics API and exposes them.

|  Metric              |  Type    |  Description                                                  |
| -------------------- | -------- | ------------------------------------------------------------- |
| `eci_signatures`     |  `gauge` | Current number of collected signatures                        |
| `eci_signature_goal` |  `gauge` | Official signature goal (1 000 000 for most initiatives)      |

---

## Quick Start

### Binary

```bash
go install github.com/tvanriel/eci-prometheus-exporter@latest
eci-prometheus-exporter \
  -initiative=045 \
  -listen-address=:8080 \
  -interval=5m
```

### Docker

```bash
docker run -d --name eci-exporter \
  -p 8080:8080 \
  docker.io/mitaka8/eci-prometheus-exporter:latest \
  -initiatives=045
```

### Kubernetes

```bash
kubectl apply -k ./deploy
```
---

## Configuration Flags

| Flag              | Default       | Description                    |
| ----------------- | ------------- | ------------------------------ |
| `-initiatives`     | **required** | Initiative IDs, e.g. `045,046` |
| `-listen-address` | `:8080`       | HTTP bind address              |
| `-interval`       | `5m`          | Polling interval               |

---

## Development

```bash
go run main.go -initiatives=045
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
