ARG REPO_BUILD_TAG="unknown"

FROM golang:1.23-alpine3.20 AS builder
ARG REPO_BUILD_TAG
COPY . /build
WORKDIR /build
RUN apk add --no-cache --update build-base \
    && CGO_ENABLED=1 go build \
        -mod=vendor -trimpath \
       -ldflags "-s -w \
            -X github.com/prometheus/common/version.Version=${REPO_BUILD_TAG} \
            -X github.com/prometheus/common/version.BuildDate=$(date +%Y-%m-%dT%H:%M:%S%z) \
            -X github.com/prometheus/common/version.Branch=$(git rev-parse --abbrev-ref HEAD) \
            -X github.com/prometheus/common/version.Revision=$(git rev-parse --short HEAD) \
            -X github.com/prometheus/common/version.BuildUser=gpbackup_exporter" \
        -o gpbackup_exporter gpbackup_exporter.go

FROM alpine:3.20
ARG REPO_BUILD_TAG
ENV TZ="Etc/UTC" \
    EXPORTER_TELEMETRY_PATH="/metrics" \
    EXPORTER_PORT="19854" \
    EXPORTER_CONFIG="" \
    COLLECT_INTERVAL="600" \
    COLLECT_DEPTH="0" \
    COLLECT_DELETED="false" \
    COLLECT_FAILED="false" \
    HISTORY_FILE="" \
    DB_INCLUDE="" \
    DB_EXCLUDE="" \
    BACKUP_TYPE=""
RUN apk add --no-cache --update \
        tzdata \
        su-exec \
        ca-certificates \
        bash \
    && cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo "${TZ}" > /etc/timezone \
    && rm -rf /var/cache/apk/*
COPY --chmod=755 docker_files/entrypoint.sh /entrypoint.sh
COPY --chmod=755 docker_files/run_exporter.sh /run_exporter.sh
COPY --from=builder --chmod=755 /build/gpbackup_exporter /gpbackup_exporter
LABEL \
    org.opencontainers.image.version="${REPO_BUILD_TAG}" \
    org.opencontainers.image.source="https://github.com/woblerr/gpbackup_exporter"
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/run_exporter.sh"]
EXPOSE ${EXPORTER_PORT}