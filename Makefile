.PHONY: # ignore

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

help:
	@perl -nle'print $& if m{^[a-zA-Z_-]+:.*?## .*$$}' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

git-push: .PHONY
	git pull --rebase
	git push

install: .PHONY
	GOBIN=~/bin go install -ldflags '-s -w $(BASE_LDFLAGS)' ./cmd/plural/

build-cli: .PHONY
	GOBIN=~/bin go build -ldflags '-s -w $(BASE_LDFLAGS)' -o $(OUTFILE) ./cmd/plural/

release: .PHONY
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags '-s -w $(BASE_LDFLAGS)'  -o plural.o ./cmd/plural/

setup: .PHONY ## sets up your local env (for mac only)
	brew install golangci-lint

plural: .PHONY ## uploads to plural
	plural apply -f plural/Pluralfile

build: .PHONY ## Build the Docker image
	docker build --build-arg APP_NAME=$(APP_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		--build-arg APP_DATE=$(APP_DATE) \
		--build-arg APP_COMMIT=$(BUILD) \
		-t $(APP_NAME):$(APP_VSN) \
		-t $(APP_NAME):latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_NAME):$(APP_VSN) .

build-cloud: .PHONY ## build the cloud docker image
	docker build --build-arg APP_NAME=$(APP_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		--build-arg APP_DATE=$(APP_DATE) \
		--build-arg APP_COMMIT=$(BUILD) \
		-t $(APP_NAME)-cloud:$(APP_VSN) \
		-t $(APP_NAME)-cloud:latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME)-cloud:$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_NAME)-cloud:$(APP_VSN) -f dockerfiles/Dockerfile.cloud  .

push: .PHONY ## push to gcr
	docker push gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN)
	docker push $(DKR_HOST)/plural/${APP_NAME}:$(APP_VSN)

push-cloud: .PHONY ## push to gcr
	docker push gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN)-cloud
	docker push $(DKR_HOST)/plural/${APP_NAME}:$(APP_VSN)-cloud

generate: .PHONY
	go generate ./...

bake-ami: .PHONY
	cd packer && packer build -var "img_name=plural/ubuntu/$(BUILD)" .
	@echo "baked ami for all regions"

up: .PHONY # spin up local server
	docker-compose up

pull: .PHONY # pulls new server image
	docker-compose pull

serve: build-cloud .PHONY # build cloud version of plural-cli and start plural serve in docker
	docker kill plural-cli || true
	docker run --rm --name plural-cli -p 8080:8080 -d plural-cli:latest-cloud

release-vsn: # tags and pushes a new release
	@read -p "Version: " tag; \
	git checkout main; \
	git pull --rebase; \
	git tag -a $$tag -m "new release"; \
	git push origin $$tag

test: .PHONY
	go test -v -race ./pkg/... ./cmd/...

format: .PHONY # formats all go code to prep for linting
	golangci-lint run --fix

genmock: .PHONY # generates mocks before running tests
	hack/gen-client-mocks.sh	

lint: .PHONY
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run

delete-tag: .PHONY
	@read -p "Version: " tag: \
	git tag -d $$tag
	git push origin :$$tag
