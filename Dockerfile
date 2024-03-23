ARG REPO_BUILD_TAG="unknown"

FROM golang:1.21-alpine3.19 AS builder
ARG REPO_BUILD_TAG
COPY . /build
WORKDIR /build
RUN apk add --no-cache --update build-base \
    && CGO_ENABLED=1 go build \
        -mod=vendor -trimpath \
        -ldflags "-X main.version=${REPO_BUILD_TAG}" \
        -o gpbackup_exporter gpbackup_exporter.go

FROM alpine:3.19
ARG REPO_BUILD_TAG
ENV EXPORTER_ENDPOINT="/metrics" \
    EXPORTER_PORT="19854" \
    EXPORTER_CONFIG="" \
    COLLECT_INTERVAL="600" \
    COLLECT_DEPTH="0" \
    HISTORY_FILE="" \
    DB_INCLUDE="" \
    DB_EXCLUDE="" \
    BACKUP_TYPE=""
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