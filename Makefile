ROOT_DIRECTORY := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

include $(ROOT_DIRECTORY)/hack/include/help.mk
include $(ROOT_DIRECTORY)/hack/include/tools.mk
include $(ROOT_DIRECTORY)/hack/include/build.mk

GCP_PROJECT ?= pluralsh
APP_NAME ?= plural-cli
APP_VSN ?= $(shell git describe --tags --always --dirty)
APP_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%S%z")
BUILD ?= $(shell git rev-parse --short HEAD)
TIMESTAMP ?= $(shell date +%s)
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
OUTFILE ?= plural.o
GOBIN ?= go env GOBIN

# Targets to run before other targets
# install-tools - Install binaries required to run targets
PRE := install-tools

.PHONY: git-push
git-push:
	git pull --rebase
	git push

.PHONY: install
install: install-cli

.PHONY: install-cli
install-cli:
	go build -ldflags '$(LDFLAGS)' -o $(GOBIN)/plural ./cmd/plural

.PHONY: build-cli
build-cli: ## Build a CLI binary for the host architecture without embedded UI
	go build -ldflags='$(LDFLAGS)' -o $(OUTFILE) ./cmd/plural

.PHONY: release
release:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='$(LDFLAGS)' -o $(OUTFILE) ./cmd/plural

.PHONY: goreleaser
goreleaser:
	goreleaser build --clean --snapshot

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

.PHONY: build-cloud-image
build-cloud-image: ## build the cloud docker image
	docker build --build-arg APP_NAME=$(APP_NAME) \
		--build-arg APP_VSN=$(APP_VSN) \
		--build-arg APP_DATE=$(APP_DATE) \
		--build-arg APP_COMMIT=$(BUILD) \
		-t $(APP_NAME)-cloud:$(APP_VSN) \
		-t $(APP_NAME)-cloud:latest \
		-t gcr.io/$(GCP_PROJECT)/$(APP_NAME)-cloud:$(APP_VSN) \
		-t $(DKR_HOST)/plural/$(APP_NAME)-cloud:$(APP_VSN) -f dockerfiles/Dockerfile.cloud  .

.PHONY: build-dind-image
build-dind-image: ## build the dind docker image
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
serve: build-cloud-image ## build cloud version of plural-cli and start plural serve in docker
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

.PHONY: e2e
e2e: --ensure-venom
	@rm -rf testout ;\
	VENOM_VAR_branch=e2e-${PLRL_CLI_E2E_PROVIDER}-${TIMESTAMP} \
	VENOM_VAR_directory=../../testout/${PLRL_CLI_E2E_PROVIDER} \
	VENOM_VAR_email=${PLRL_CLI_E2E_SA_EMAIL} \
	VENOM_VAR_gitRepo=${PLRL_CLI_E2E_GIT_REPO} \
	VENOM_VAR_gitRepoPrivateKeyPath=${PLRL_CLI_E2E_PRIVATE_KEY_PATH} \
	VENOM_VAR_username=${PLRL_CLI_E2E_SA_USERNAME} \
	VENOM_VAR_token=${PLRL_CLI_E2E_SA_TOKEN} \
	VENOM_VAR_pluralHome=${HOME}/.plural \
	VENOM_VAR_pluralKey=${PLRL_CLI_E2E_PLURAL_KEY} \
	VENOM_VAR_project=${PLRL_CLI_E2E_PROJECT}-${TIMESTAMP} \
	VENOM_VAR_provider=${PLRL_CLI_E2E_PROVIDER} \
	VENOM_VAR_region=${PLRL_CLI_E2E_REGION} \
	VENOM_VAR_azureSubscriptionId=${PLRL_CLI_E2E_AZURE_SUBSCRIPTION_ID} \
	VENOM_VAR_azureTenantId=${PLRL_CLI_E2E_AZURE_TENANT_ID} \
	VENOM_VAR_azureStorageAccount=${PLRL_CLI_E2E_AZURE_STORAGE_ACCOUNT}${TIMESTAMP} \
	VENOM_VAR_gcpOrgID=${PLRL_CLI_E2E_GCLOUD_ORG_ID} \
	VENOM_VAR_gcpBillingID=${PLRL_CLI_E2E_GCLOUD_BILLING_ID} \
	VENOM_VAR_awsZoneA=${PLRL_CLI_E2E_AWS_ZONE_A} \
	VENOM_VAR_awsZoneB=${PLRL_CLI_E2E_AWS_ZONE_B} \
	VENOM_VAR_awsZoneC=${PLRL_CLI_E2E_AWS_ZONE_C} \
	VENOM_VAR_awsProject=${PLRL_CLI_E2E_PROJECT} \
	VENOM_VAR_awsBucket=e2e-tf-state-${TIMESTAMP} \
	PLURAL_LOGIN_AFFIRM_CURRENT_USER=true \
	PLURAL_UP_AFFIRM_DEPLOY=true \
	PLURAL_DOWN_AFFIRM_DESTROY=true \
	PLURAL_CD_USE_EXISTING_CREDENTIALS=true \
	TF_VAR_network=plural-e2e-network-${TIMESTAMP} \
	TF_VAR_subnetwork=plural-e2e-subnet-${TIMESTAMP} \
	TF_VAR_deletion_protection=false \
 		venom run -vv --html-report --format=json --output-dir testout test/plural

.PHONY: format
format: ## formats all go code to prep for linting
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.62.2 golangci-lint run --fix

.PHONY: genmock
genmock: ## generates mocks before running tests
	hack/gen-client-mocks.sh

.PHONY: lint
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.62.2 golangci-lint run

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
