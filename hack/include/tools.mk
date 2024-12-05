VENOM_BINARY := $(shell which venom)
VENOM_VERSION := v1.2.0

GO_BINARY := $(shell which go)
GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MIN_GO_MAJOR_VERSION = 1
MIN_GO_MINOR_VERSION = 22

.PHONY: install-tools
install-tools: --ensure-go --ensure-venom ## Install required dependencies to run make targets

.PHONY: --ensure-go
--ensure-go:
ifndef GO_BINARY
	$(error "Cannot find go binary")
endif
	@if [ $(GO_MAJOR_VERSION) -gt $(MIN_GO_MAJOR_VERSION) ]; then \
		exit 0 ;\
  elif [ $(GO_MAJOR_VERSION) -lt $(MIN_GO_MAJOR_VERSION) ]; then \
		exit 1; \
  elif [ $(GO_MINOR_VERSION) -lt $(MIN_GO_MINOR_VERSION) ] ; then \
		exit 1; \
  fi

.PHONY: --ensure-venom
--ensure-venom:
ifndef VENOM_BINARY
	@echo "[tools] downloading venom..."
	@go install github.com/ovh/venom/cmd/venom@${VENOM_VERSION}
else
	@echo "[tools] venom already exists"
endif
