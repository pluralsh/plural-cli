WAILS_BINARY := $(shell which wails)
WAILS_VERSION := v2.4.1

GO_BINARY := $(shell which go)
GO_MAJOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MIN_GO_MAJOR_VERSION = 1
MIN_GO_MINOR_VERSION = 19

.PHONY: install-tools
install-tools: --ensure-go --ensure-wails ## Install required dependencies to run make targets

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

.PHONY: --ensure-wails
--ensure-wails:
ifndef WAILS_BINARY
	@echo "[tools] downloading wails..."
	@go install github.com/wailsapp/wails/v2/cmd/wails@$(WAILS_VERSION)
else
	@echo "[tools] wails already exists"
endif