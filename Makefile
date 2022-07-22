VERSION=$(shell git tag | tail -n 1)
DATE=$(shell date --iso-8601=seconds)
CMD=vpnlist
BUILD_DIR=./bin

.PHONY: build version
build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X 'main.BuildTime=$(DATE)' -X 'main.Version=$(VERSION)'" -o $(BUILD_DIR)/$(CMD) .
version:
	@echo $(VERSION)
