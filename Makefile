#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
APPNAME := nexus
BINARY := nexusd

# Build tags
build_tags = netgo
build_tags := $(strip $(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(APPNAME) \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(BINARY) \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                  Build                                  ###
###############################################################################

all: install

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/$(BINARY)

build:
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(BINARY) ./cmd/$(BINARY)

build-linux:
	GOOS=linux GOARCH=amd64 go build -mod=readonly $(BUILD_FLAGS) -o build/$(BINARY)-linux-amd64 ./cmd/$(BINARY)

go.sum: go.mod
	@echo "Ensuring dependencies have not been modified..."
	go mod verify
	go mod tidy

clean:
	rm -rf build/

###############################################################################
###                                  Proto                                  ###
###############################################################################

proto-gen:
	@echo "Generating Protobuf files..."
	@./scripts/protocgen.sh

###############################################################################
###                               Initialize                                ###
###############################################################################

init:
	./scripts/init.sh

###############################################################################
###                                 Testing                                 ###
###############################################################################

test:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

###############################################################################
###                                 Linting                                 ###
###############################################################################

lint:
	golangci-lint run --out-format=tab

.PHONY: all install build build-linux clean proto-gen init test test-cover lint
