ARG REPO_BUILD_TAG="unknown"

FROM golang:1.18-alpine AS builder
ARG REPO_BUILD_TAG
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 go build \
        -mod=vendor -trimpath \
        -ldflags "-X main.version=${REPO_BUILD_TAG}" \
        -o gpbackup_exporter gpbackup_exporter.go

FROM alpine:3.17
ARG REPO_BUILD_TAG
ENV EXPORTER_ENDPOINT="/metrics" \
    EXPORTER_PORT="19854" \
    COLLECT_INTERVAL="600" \
    HISTORY_FILE=""
RUN apk add --no-cache --update \
        ca-certificates \
        bash \
    && rm -rf /var/cache/apk/*
COPY --chmod=755 docker_files/run_exporter.sh /run_exporter.sh
COPY --from=builder --chmod=755 /build/gpbackup_exporter /gpbackup_exporter
USER nobody
LABEL \
    org.opencontainers.image.version="${REPO_BUILD_TAG}" \
    org.opencontainers.image.source="https://github.com/woblerr/gpbackup_exporter"
ENTRYPOINT ["/run_exporter.sh"]
EXPOSE ${EXPORTER_PORT}