# AGENTS.md

## Project Overview

`gpbackup_exporter` is a Prometheus exporter that collects metrics from the SQLite `gpbackup_history.db` file and exposes them over HTTP.
The main entrypoint, `gpbackup_exporter.go`, defines the Kingpin CLI, configures logging and the exporter-toolkit web server, and schedules collection.

The exporter can run directly on the Greenplum master host with access to the history database, as a systemd service, or in the Docker image built by this repository.

## Project Structure

* `gpbackup_exporter.go` contains CLI flags, application startup, signal handling, and the collection loop.
* `gpbckpexporter/` contains history-database parsing, collection behavior, metric definitions, converters, HTTP endpoint setup, and unit tests.
* `docker_files/run_exporter.sh` maps Docker environment variables to CLI flags.
* `e2e_tests/` contains Docker-based HTTP, TLS, basic-auth, and client-certificate test fixtures and assertions.
* `Dockerfile`, `Dockerfile.artifacts`, and `Makefile` are the main build surfaces.
* `.github/workflows/build.yml` runs unit tests, lint, e2e tests, image publishing, and GoReleaser.
* `vendor/` is committed dependency source; do not edit vendored files by hand.

## Build, Test, and Lint

Prefer existing Makefile targets over equivalent raw commands.
The project and CI use Go 1.25 and vendored modules.

* Run unit tests with `make test`.
  The target runs `TZ="Etc/UTC" go test -mod=vendor -timeout=60s -count 1 ./...`.
* For focused iteration, preserve the same flags and select a package and test, for example `TZ="Etc/UTC" go test -mod=vendor -timeout=60s -count 1 ./gpbckpexporter -run TestName`.
* `make build` runs unit tests and builds a Linux AMD64 binary.
* `make build-on-darwin` runs unit tests and cross-builds a static Linux AMD64 binary; it requires `musl-cross`.
* `make build-darwin` runs unit tests and builds a Darwin AMD64 binary.
* `make docker` builds the default Docker image.
* `make dist` builds snapshot archives and packages through Docker and GoReleaser, then writes them under the ignored `dist/` directory.
* Run lint with `make lint`.
* Format Go files with `make fmt`.

Every Make invocation evaluates `docker ps` while parsing the Makefile, including `make -n` and targets that otherwise only run Go commands.
A Docker daemon warning from that lookup does not mean e2e tests ran.

## End-to-End Tests

Run the complete suite with `make test-e2e`.
The target requires Docker, pulls and builds an image, runs HTTP/TLS/basic-auth/client-certificate cases, and repeatedly force-removes the `gpbackup_exporter_e2e` container.
It may first remove a pre-existing container matched by that name.
Do not run `make test-e2e` unless the user explicitly requests it or approves it.

The certificates and private keys under `e2e_tests/` are committed test fixtures only.
Never reuse them for real services or replace them with real credentials.
Keep `e2e_tests/README.md`, e2e scripts and fixtures, and Makefile e2e cases synchronized when e2e behavior or coverage changes.

## Verification

* For Go changes, run `make fmt`, `make test`, and `make lint`.
* Focused tests are useful during iteration, but do not use them as the only final verification.
* For build, Docker, or packaging changes, run the relevant Makefile target when practical.
* Run `make test-e2e` only under the approval rules described above.
* For documentation-only changes, do not run code tests unless needed to validate a documented command.
* Always run `git diff --check` before handing off changes.

## Code Style and Test Patterns

* Keep CLI definition and process lifecycle in `gpbackup_exporter.go`.
* Keep history-database parsing, collection behavior, metric logic, and HTTP endpoint code in `gpbckpexporter/`.
* Prefer simple, concrete changes within the existing package boundaries.
  Add an abstraction, interface, or wrapper only when it removes real duplication, reduces existing complexity, or provides a focused test seam.
* Follow the existing standard-library test style: `testing`, table-driven cases, `t.Run`, and direct fatal or error assertions.
* Metric tests use an isolated `prometheus.NewRegistry()` and reset package-level metrics where needed to prevent state from leaking between cases.
* Add focused regression coverage for changed parsing, filters, labels, metric values, and error behavior.
* Do not introduce `testify` or another test framework solely for convenience.
* For dependency changes, update `go.mod` and `go.sum`, then regenerate committed dependency sources with `go mod vendor`.

## Keep Synchronized

* When CLI flags change, update their Kingpin definitions, relevant unit tests, README flag documentation and examples, and, when exposed in Docker, `Dockerfile`, `docker_files/run_exporter.sh`, and relevant e2e coverage.
* When Docker environment variables change, keep `Dockerfile`, `docker_files/run_exporter.sh`, README environment-variable documentation, and relevant e2e coverage aligned.
* When metrics or filtering semantics change, update the definitions and tests under `gpbckpexporter/`, README metric tables and behavior notes, and e2e metric assertions when they cover the changed behavior.
* Keep the Go version aligned across `go.mod`, `Dockerfile`, and `.github/workflows/build.yml`.
* Keep the GoReleaser version aligned between `Dockerfile.artifacts` and `.github/workflows/build.yml`.

## Runtime and Safety Rules

* `make prepare-service` overwrites the ignored local `gpbackup_exporter.service` from the template and substitutes the current repository path and `HISTORY_FILE` value.
  Review the generated unit before installing it.
* `make install-service` writes to `/etc/systemd/system`, reloads systemd, enables the exporter, and restarts it.
* `make remove-service` stops and disables the exporter, removes its systemd unit, and reloads systemd.
* Never run `make install-service` or `make remove-service` without explicit user approval.
* Do not publish images, packages, tags, or releases unless the user explicitly requests it.

## Generated, Local, and Release Artifacts

* Do not commit ignored local artifacts such as the built `gpbackup_exporter` binary, `gpbackup_exporter.service`, `dist/`, `.vscode/`, `.DS_Store`, or coverage `*.out` and `*.html` files.
* Keep the committed `vendor/` directory consistent with `go.mod` and `go.sum`.
* GoReleaser has changelog generation disabled; do not add changelog work unless a release task explicitly requires it.

## Editing These Notes

Keep this file concise, operational, and limited to repository-specific facts.
Use semantic line breaks and keep one sentence per line where practical so future changes produce small, reviewable diffs.
