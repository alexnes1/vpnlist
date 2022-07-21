VERSION=$(shell git tag | tail -n 1)
DATE=$(shell date --iso-8601=seconds)
CMD=vpnlist

.PHONY: build version
build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X 'main.BuildTime=$(DATE)' -X 'main.Version=$(VERSION)'" -o ./bin/$(CMD) .
version:
	@echo $(VERSION)
