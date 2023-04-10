ROOT_DIRECTORY := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

include $(ROOT_DIRECTORY)/hack/include/help.mk
include $(ROOT_DIRECTORY)/hack/include/tools.mk
include $(ROOT_DIRECTORY)/hack/include/build.mk

GCP_PROJECT ?= pluralsh
APP_NAME ?= plural-cli
APP_VSN ?= $(shell git describe --tags --always --dirty)
APP_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%S%z")
BUILD ?= $(shell git rev-parse --short HEAD)
DKR_HOST ?= dkr.plural.sh
GOOS ?= darwin
GOARCH ?= arm64
GOLANG_CROSS_VERSION  ?= v1.20.2
PACKAGE ?= github.com/pluralsh/plural
BASE_LDFLAGS ?= -s -w
LDFLAGS ?= $(BASE_LDFLAGS) $\
	-X "$(PACKAGE)/cmd/plural.Version=$(APP_VSN)" $\
	-X "$(PACKAGE)/cmd/plural.Commit=$(BUILD)" $\
	-X "$(PACKAGE)/cmd/plural.Date=$(APP_DATE)" $\
	-X "$(PACKAGE)/pkg/scm.GitlabClientSecret=${GITLAB_CLIENT_SECRET}" $\
	-X "$(PACKAGE)/pkg/scm.BitbucketClientSecret=${BITBUCKET_CLIENT_SECRET}"
WAILS_TAGS ?= desktop,production,ui
WAILS_BINDINGS_TAGS ?= bindings,generate
WAILS_BINDINGS_BINARY_NAME ?= wailsbindings
TAGS ?= $(WAILS_TAGS)
OUTFILE ?= plural.o

# Targets to run before other targets
# install-tools - Install binaries required to run targets
PRE := install-tools

.PHONY: git-push
git-push:
	git pull --rebase
	git push

.PHONY: install
install:
	go install -ldflags '$(LDFLAGS)' .

.PHONY: build-cli
build-cli: ## Build a CLI binary for the host architecture without embedded UI
	go build -ldflags='$(LDFLAGS)' -o $(OUTFILE) .

.PHONY: build-cli-ui
build-cli-ui: $(PRE) generate-bindings ## Build a CLI binary for the host architecture with embedded UI
	CGO_LDFLAGS=$(CGO_LDFLAGS) go build -tags $(WAILS_TAGS) -ldflags='$(LDFLAGS)' -o $(OUTFILE) .

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
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='$(LDFLAGS)' -o $(OUTFILE) .

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
	cd packer && packer build -var "img_name=plural/ubuntu/$(BUILD)" .
	@echo "baked ami for all regions"

.PHONY: up
up: # spin up local server
	docker-compose up

.PHONY: pull
pull: # pulls new server image
	docker-compose pull

.PHONY: serve
serve: build-cloud # build cloud version of plural-cli and start plural serve in docker
	docker kill plural-cli || true
	docker run --rm --name plural-cli -p 8080:8080 -d plural-cli:latest-cloud

.PHONY: release-vsn
release-vsn: # tags and pushes a new release
	@read -p "Version: " tag; \
	git checkout main; \
	git pull --rebase; \
	git tag -a $$tag -m "new release"; \
	git push origin $$tag

.PHONY: test
test:
	go test -v -race ./pkg/... ./cmd/...

.PHONY: format
format: # formats all go code to prep for linting
	golangci-lint run --fix

.PHONY: genmock
genmock: # generates mocks before running tests
	hack/gen-client-mocks.sh	

.PHONY: lint
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run

.PHONY: delete-tag
delete-tag:
	@read -p "Version: " tag: \
	git tag -d $$tag
	git push origin :$$tag
