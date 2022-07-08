.PHONY: # ignore

GCP_PROJECT ?= pluralsh
APP_NAME ?= plural-cli
APP_VSN ?= $(shell cat VERSION)
BUILD ?= $(shell git rev-parse --short HEAD)
DKR_HOST ?= dkr.plural.sh
GOOS ?= darwin
GOARCH ?= amd64
BASE_LDFLAGS ?= -X main.GitCommit=$(BUILD) -X main.Version=$(APP_VSN) -X github.com/pluralsh/plural/pkg/scm.GitlabClientSecret=${GITLAB_CLIENT_SECRET}
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

plural: .PHONY ## uploads to plural
	plural apply

build: .PHONY ## Build the Docker image
	docker build --build-arg APP_NAME=$(APP_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		-t $(APP_NAME):$(APP_VSN) \
		-t $(APP_NAME):latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_NAME):$(APP_VSN) .

build-cloud: .PHONY ## build the cloud docker image
	docker build --platform linux/amd64 -t $(APP_NAME):$(APP_VSN)-cloud \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN)-cloud \
		-t $(DKR_HOST)/plural/$(APP_NAME):$(APP_VSN)-cloud -f dockerfiles/Dockerfile.cloud  .

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

release-vsn: # tags and pushes a new release
	@read -p "Version: " tag; \
	git tag -a $$tag -m "new release"; \
	git push origin $$tag

test: .PHONY
	go test -v -race ./pkg/... ./cmd/...
