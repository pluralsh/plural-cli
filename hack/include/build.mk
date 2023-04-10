# BUILDOS is the host machine OS
BUILDOS ?= $(shell uname -s)

# CGO_LDFLAGS is required when building on darwin
CGO_LDFLAGS ?= ""

ifeq ($(BUILDOS),Darwin)
	CGO_LDFLAGS="-framework UniformTypeIdentifiers"
endif
