ROOT_DIRECTORY := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

include $(ROOT_DIRECTORY)/hack/include/help.mk
include $(ROOT_DIRECTORY)/hack/include/tools.mk
include $(ROOT_DIRECTORY)/hack/include/build.mk

GCP_PROJECT ?= pluralsh
APP_NAME ?= plural-cli
APP_CTL_NAME ?= plrlctl
APP_VSN ?= $(shell git describe --tags --always --dirty)
APP_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%S%z")
BUILD ?= $(shell git rev-parse --short HEAD)
DKR_HOST ?= dkr.plural.sh
GOOS ?= darwin
GOARCH ?= arm64
GOLANG_CROSS_VERSION  ?= v1.22.0
PACKAGE ?= github.com/pluralsh/plural-cli
BASE_LDFLAGS ?= -s -w
LDFLAGS ?= $(BASE_LDFLAGS) $\
	-X "$(PACKAGE)/pkg/common.Version=$(APP_VSN)" $\
	-X "$(PACKAGE)/pkg/common.Commit=$(BUILD)" $\
	-X "$(PACKAGE)/pkg/common.Date=$(APP_DATE)" $\
	-X "$(PACKAGE)/pkg/scm.GitlabClientSecret=${GITLAB_CLIENT_SECRET}" $\
	-X "$(PACKAGE)/pkg/scm.BitbucketClientSecret=${BITBUCKET_CLIENT_SECRET}"
WAILS_TAGS ?= desktop,production,ui,debug
WAILS_BINDINGS_TAGS ?= bindings,generate
WAILS_BINDINGS_BINARY_NAME ?= wailsbindings
TAGS ?= $(WAILS_TAGS)
OUTFILE ?= plural.o
OUTCTLFILE ?= plrlctl.o
GOBIN ?= go env GOBIN

# Targets to run before other targets
# install-tools - Install binaries required to run targets
PRE := install-tools

.PHONY: git-push
git-push:
	git pull --rebase
	git push

.PHONY: install
install:
	go build -ldflags '$(LDFLAGS)' -o $(GOBIN)/plural ./cmd/plural
	go build -ldflags '$(LDFLAGS)' -o $(GOBIN)/plrlctl ./cmd/plrlctl

.PHONY: build-cli
build-cli: ## Build a CLI binary for the host architecture without embedded UI
	go build -ldflags='$(LDFLAGS)' -o $(OUTFILE) ./cmd/plural

.PHONY: build-ctl
build-ctl: ## Build a CLI binary for the fleet management
	go build -ldflags='$(LDFLAGS)' -o $(OUTCTLFILE) ./cmd/plrlctl

.PHONY: build-cli-ui
build-cli-ui: $(PRE) generate-bindings ## Build a CLI binary for the host architecture with embedded UI
	CGO_LDFLAGS=$(CGO_LDFLAGS) go build -tags $(WAILS_TAGS) -ldflags='$(LDFLAGS)' -o $(OUTFILE) ./cmd/plural

.PHONY: build-web
build-web: ## Build just the embedded UI
	cd pkg/ui/web && yarn --immutable && yarn build

.PHONY: run-web
run-web: $(PRE) ## Run the UI for development
	@CGO_LDFLAGS=$(CGO_LDFLAGS) wails dev -tags ui -browser -skipbindings

# This is somewhat an equivalent of wails `GenerateBindings` method.
# Ref: https://github.com/wailsapp/wails/blob/master/v2/pkg/commands/bindings/bindings.go#L28
.PHONY: generate-bindings
generate-bindings: build-web ## Generate backend bindings for the embedded UI
	@echo Building bindings binary
	@CGO_LDFLAGS=$(CGO_LDFLAGS) go build -tags $(WAILS_BINDINGS_TAGS) -ldflags='$(LDFLAGS)' -o $(WAILS_BINDINGS_BINARY_NAME) .
	@echo Generating bindings
	@./$(WAILS_BINDINGS_BINARY_NAME) > /dev/null 2>&1
	@echo Cleaning up
	@rm $(WAILS_BINDINGS_BINARY_NAME)

.PHONY: release
release:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='$(LDFLAGS)' -o $(OUTFILE) ./cmd/plural
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='$(LDFLAGS)' -o $(OUTCTLFILE) ./cmd/plrlctl

.PHONY: goreleaser
goreleaser:
	goreleaser release --clean --prepare --snapshot

.PHONY: setup
setup: ## sets up your local env (for mac only)
	brew install golangci-lint

.PHONY: plural
plural: ## uploads to plural
	plural apply -f plural/Pluralfile

.PHONY: build
build: ## Build the Docker image
	docker build --build-arg APP_NAME=$(APP_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		--build-arg APP_DATE=$(APP_DATE) \
		--build-arg APP_COMMIT=$(BUILD) \
		-t $(APP_NAME):$(APP_VSN) \
		-t $(APP_NAME):latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_NAME):$(APP_VSN) .

.PHONY: build-ctl
build-ctl: ## Build the plrctl Docker image
	docker build --build-arg APP_NAME=$(APP_CTL_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		--build-arg APP_DATE=$(APP_DATE) \
		--build-arg APP_COMMIT=$(BUILD) \
		-t $(APP_CTL_NAME):$(APP_VSN) \
		-t $(APP_CTL_NAME):latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_CTL_NAME):$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_CTL_NAME):$(APP_VSN) -f dockerfiles/plrlctl/Dockerfile  .

.PHONY: build-cloud
build-cloud: ## build the cloud docker image
	docker build --build-arg APP_NAME=$(APP_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		--build-arg APP_DATE=$(APP_DATE) \
		--build-arg APP_COMMIT=$(BUILD) \
		-t $(APP_NAME)-cloud:$(APP_VSN) \
		-t $(APP_NAME)-cloud:latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME)-cloud:$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_NAME)-cloud:$(APP_VSN) -f dockerfiles/Dockerfile.cloud  .

.PHONY: build-dind
build-dind: ## build the dind docker image
	docker build --build-arg APP_NAME=$(APP_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		--build-arg APP_DATE=$(APP_DATE) \
		--build-arg APP_COMMIT=$(BUILD) \
		-t $(APP_NAME)-cloud:$(APP_VSN) \
		-t $(APP_NAME)-cloud:latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME)-cloud:$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_NAME)-dind:$(APP_VSN) -f dockerfiles/Dockerfile.dind  .

.PHONY: push
push: ## push to gcr
	docker push gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN)
	docker push $(DKR_HOST)/plural/${APP_NAME}:$(APP_VSN)

.PHONY: push-cloud
push-cloud: ## push to gcr
	docker push gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN)-cloud
	docker push $(DKR_HOST)/plural/${APP_NAME}:$(APP_VSN)-cloud

.PHONY: generate
generate:
	go generate ./...

.PHONY: bake-ami
bake-ami:
	cd packer && packer build -var "cli_version=$(APP_VSN)" .
	@echo "baked ami for all regions"

.PHONY: up
up: ## spin up local server
	docker-compose up

.PHONY: pull
pull: ## pulls new server image
	docker-compose pull

.PHONY: serve
serve: build-cloud ## build cloud version of plural-cli and start plural serve in docker
	docker kill plural-cli || true
	docker run --rm --name plural-cli -p 8080:8080 -d plural-cli:latest-cloud

.PHONY: release-vsn
release-vsn: ## tags and pushes a new release
	@read -p "Version: " tag; \
	git checkout main; \
	git pull --rebase; \
	git tag -a $$tag -m "new release"; \
	git push origin $$tag

.PHONY: setup-tests
setup-tests:
	go install gotest.tools/gotestsum@latest

.PHONY: test
test: setup-tests
	gotestsum --format testname -- -v -race ./pkg/... ./cmd/command/...

.PHONY: format
format: ## formats all go code to prep for linting
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run --fix

.PHONY: genmock
genmock: ## generates mocks before running tests
	hack/gen-client-mocks.sh

.PHONY: lint
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run

.PHONY: delete-tag
delete-tag:
	@read -p "Version: " tag; \
	git tag -d $$tag; \
	git push origin :$$tag

REPO_URL := https://github.com/pluralsh/plural-cli/releases/download
OIDC_ISSUER_URL := https://token.actions.githubusercontent.com
VERIFY_FILE_NAME := checksums.txt
RELEASE_ARCHIVE_NAME := plural-cli
VERIFY_TMP_DIR := dist
PUBLIC_KEY_FILE := cosign.pub

.PHONY: verify
verify: ## verifies provided tagged release with cosign
	@read -p "Enter version to verify: " tag ;\
	echo "Downloading ${VERIFY_FILE_NAME} for tag v$${tag}..." ;\
	wget -P ${VERIFY_TMP_DIR} "${REPO_URL}/v$${tag}/checksums.txt" >/dev/null 2>&1 ;\
	echo "Verifying signature..." ;\
	cosign verify-blob \
	  --key "${PUBLIC_KEY_FILE}" \
      --signature "${REPO_URL}/v$${tag}/${VERIFY_FILE_NAME}.sig" \
      "./${VERIFY_TMP_DIR}/${VERIFY_FILE_NAME}" ;\
    echo "Verifying archives..." ;\
    wget -P ${VERIFY_TMP_DIR} "${REPO_URL}/v$${tag}/${RELEASE_ARCHIVE_NAME}_$${tag}_Darwin_amd64.tar.gz" >/dev/null 2>&1 ;\
    wget -P ${VERIFY_TMP_DIR} "${REPO_URL}/v$${tag}/${RELEASE_ARCHIVE_NAME}_$${tag}_Darwin_arm64.tar.gz" >/dev/null 2>&1 ;\
    wget -P ${VERIFY_TMP_DIR} "${REPO_URL}/v$${tag}/${RELEASE_ARCHIVE_NAME}_$${tag}_Linux_amd64.tar.gz" >/dev/null 2>&1 ;\
    wget -P ${VERIFY_TMP_DIR} "${REPO_URL}/v$${tag}/${RELEASE_ARCHIVE_NAME}_$${tag}_Linux_arm64.tar.gz" >/dev/null 2>&1 ;\
    wget -P ${VERIFY_TMP_DIR} "${REPO_URL}/v$${tag}/${RELEASE_ARCHIVE_NAME}_$${tag}_Windows_amd64.tar.gz" >/dev/null 2>&1 ;\
    (cd ${VERIFY_TMP_DIR} && exec sha256sum --ignore-missing -c checksums.txt) ;\
    rm -r "${VERIFY_TMP_DIR}"
