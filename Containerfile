FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS builder

ENV DEBIAN_FRONTEND=noninteractive

COPY go.mod go.sum ./
RUN go mod download

COPY . /usr/src/eci-prometheus-exporter
WORKDIR /usr/src/eci-prometheus-exporter

ARG TARGETOS TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /usr/bin/eci-prometheus-exporter .

FROM debian:stable-slim AS final
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /opt/eci-prometheus-exporter
RUN apt-get update && apt-get install -y ca-certificates && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/bin/eci-prometheus-exporter /usr/bin/eci-prometheus-exporter

CMD ["/usr/bin/eci-prometheus-exporter"]
