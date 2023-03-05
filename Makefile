SHELL := /bin/bash
APP_NAME := gpbackup_exporter
BRANCH_FULL := $(shell git rev-parse --abbrev-ref HEAD)
BRANCH := $(subst /,-,$(BRANCH_FULL))
GIT_REV := $(shell git describe --abbrev=7 --always)
SERVICE_CONF_DIR := /etc/systemd/system
HISTORY_FILE ?= /data/master/gpseg-1/gpbackup_history.yaml
HTTP_PORT := 19854
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
DOCKER_CONTAINER_E2E := $(shell docker ps -a -q -f name=$(APP_NAME)_e2e)
HTTP_PORT_E2E := $(shell echo $$((10000 + ($$RANDOM % 10000))))

.PHONY: test
test:
	@echo "Run tests for $(APP_NAME)"
	TZ="Etc/UTC" go test -mod=vendor -timeout=60s -count 1  ./...

.PHONY: test-e2e
test-e2e:
	@echo "Run end-to-end tests for $(APP_NAME)"
	@if [ -n "$(DOCKER_CONTAINER_E2E)" ]; then docker rm -f "$(DOCKER_CONTAINER_E2E)"; fi;
	DOCKER_BUILDKIT=1 docker build --pull -f Dockerfile --build-arg REPO_BUILD_TAG=$(BRANCH)-$(GIT_REV) -t $(APP_NAME)_e2e .
	$(call e2e_basic,$(PWD)/e2e_tests/test_data/:/e2e_tests/:ro,/e2e_tests/gpbackup_history.yaml)
	$(call e2e_tls_auth,$(PWD)/e2e_tests/test_data/:/e2e_tests/:ro,/e2e_tests/gpbackup_history.yaml,/e2e_tests/web_config_empty.yml,false,false)
	$(call e2e_tls_auth,$(PWD)/e2e_tests/test_data/:/e2e_tests/:ro,/e2e_tests/gpbackup_history.yaml,/e2e_tests/web_config_TLS_noAuth.yml,true,false)
	$(call e2e_tls_auth,$(PWD)/e2e_tests/test_data/:/e2e_tests/:ro,/e2e_tests/gpbackup_history.yaml,/e2e_tests/web_config_TLS_Auth.yml,true,true)
	$(call e2e_tls_auth,$(PWD)/e2e_tests/test_data/:/e2e_tests/:ro,/e2e_tests/gpbackup_history.yaml,/e2e_tests/web_config_noTLS_Auth.yml,false,true)

.PHONY: build
build:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -trimpath -ldflags "-X main.version=$(BRANCH)-$(GIT_REV)" -o $(APP_NAME) $(APP_NAME).go

.PHONY: build-arm
build-arm:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -mod=vendor -trimpath -ldflags "-X main.version=$(BRANCH)-$(GIT_REV)" -o $(APP_NAME) $(APP_NAME).go

.PHONY: build-darwin
build-darwin:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -mod=vendor -trimpath -ldflags "-X main.version=$(BRANCH)-$(GIT_REV)" -o $(APP_NAME) $(APP_NAME).go

.PHONY: dist
dist:
	- @mkdir -p dist
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.artifacts --progress=plain -t gpbackup_exporter_dist .
	- @docker rm -f gpbackup_exporter_dist 2>/dev/null || exit 0
	docker run -d --name=gpbackup_exporter_dist gpbackup_exporter_dist
	docker cp gpbackup_exporter_dist:/artifacts dist/
	docker rm -f gpbackup_exporter_dist

.PHONY: docker
docker:
	@echo "Build $(APP_NAME) docker container"
	@echo "Version $(BRANCH)-$(GIT_REV)"
	DOCKER_BUILDKIT=1 docker build --pull -f Dockerfile --build-arg REPO_BUILD_TAG=$(BRANCH)-$(GIT_REV) -t $(APP_NAME) .

.PHONY: prepare-service
prepare-service:
	@echo "Prepare config file $(APP_NAME).service for systemd"
	cp $(ROOT_DIR)/$(APP_NAME).service.template $(ROOT_DIR)/$(APP_NAME).service
	sed -i.bak "s|/usr/bin|$(ROOT_DIR)|g" $(APP_NAME).service
	sed -i.bak "s|/data/master/gpseg-1/gpbackup_history.yaml|$(HISTORY_FILE)|g" $(APP_NAME).service
	rm $(APP_NAME).service.bak

.PHONY: install-service
install-service:
	@echo "Install $(APP_NAME) as systemd service"
	$(call service-install)

.PHONY: remove-service
remove-service:
	@echo "Delete $(APP_NAME) systemd service"
	$(call service-remove)

define service-install
	cp $(ROOT_DIR)/$(APP_NAME).service $(SERVICE_CONF_DIR)/$(APP_NAME).service
	systemctl daemon-reload
	systemctl enable $(APP_NAME)
	systemctl restart $(APP_NAME)
	systemctl status $(APP_NAME)
endef

define service-remove
	systemctl stop $(APP_NAME)
	systemctl disable $(APP_NAME)
	rm $(SERVICE_CONF_DIR)/$(APP_NAME).service
	systemctl daemon-reload
	systemctl reset-failed
endef

define e2e_basic
	docker run -d -p $(HTTP_PORT_E2E):$(HTTP_PORT) -v "${1}" --env HISTORY_FILE="${2}" --name=$(APP_NAME)_e2e $(APP_NAME)_e2e
	@sleep 30
	$(ROOT_DIR)/e2e_tests/run_e2e.sh $(HTTP_PORT_E2E)
	docker rm -f $(APP_NAME)_e2e
endef

define e2e_tls_auth
	docker run -d -p $(HTTP_PORT_E2E):$(HTTP_PORT)  -v "${1}" --env HISTORY_FILE="${2}"  --env EXPORTER_CONFIG="${3}" --name=$(APP_NAME)_e2e $(APP_NAME)_e2e
	@sleep 30
	$(ROOT_DIR)/e2e_tests/run_e2e.sh $(HTTP_PORT_E2E) ${4} ${5}
    docker rm -f $(APP_NAME)_e2e
endef
