.PHONY: # ignore

GCP_PROJECT ?= pluralsh
APP_NAME ?= plural-cli
APP_VSN ?= `cat VERSION`
BUILD ?= `git rev-parse --short HEAD`
DKR_HOST ?= dkr.plural.sh
GOOS ?= darwin
GOARCH ?= amd64

help:
	@perl -nle'print $& if m{^[a-zA-Z_-]+:.*?## .*$$}' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install:
	GOBIN=/usr/local/bin go install -ldflags '-s -w' ./cmd/plural/

release:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags '-s -w'  -o plural.o ./cmd/plural/

plural: .PHONY ## uploads to plural
	plural apply

build: .PHONY ## Build the Docker image
	docker build --build-arg APP_NAME=$(APP_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		-t $(APP_NAME):$(APP_VSN) \
		-t $(APP_NAME):latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_NAME):$(APP_VSN) .

push: ## push to gcr
	docker push gcr.io/$(GCP_PROJECT)/$(APP_NAME):$(APP_VSN)
	docker push $(DKR_HOST)/plural/${APP_NAME}:$(APP_VSN)

generate:
	go generate ./...