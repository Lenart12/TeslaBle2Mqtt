# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN echo "Building on $BUILDPLATFORM for $TARGETPLATFORM"

WORKDIR /build
COPY . .

# Set the target platform architecture
RUN case "$TARGETPLATFORM" in \
    "linux/amd64") GOARCH=amd64 ;; \
    "linux/arm64") GOARCH=arm64 ;; \
    "linux/arm/v7") GOARCH=arm GOARM=7 ;; \
    *) echo "Unsupported platform: $TARGETPLATFORM" && exit 1 ;; \
    esac && \
    GOOS=linux CGO_ENABLED=0 GOARCH=$GOARCH GOARM=$GOARM go build -o teslable2mqtt

FROM alpine:latest

COPY --from=builder /build/teslable2mqtt /usr/local/bin/

ENTRYPOINT ["teslable2mqtt"]

LABEL org.opencontainers.image.description="Tesla BLE to MQTT bridge"
LABEL org.opencontainers.image.source="https://github.com/Lenart12/TeslaBle2Mqtt"
