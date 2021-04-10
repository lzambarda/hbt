NAME:=hbtsrv
BUILD_TAG:=$(shell git describe --tags)
BUILDFLAGS:="-s -w -X internal.Version=$(BUILD_TAG)"

.PHONY: build
build:
	go build -ldflags $(BUILDFLAGS) -o ./bin/$(NAME) ./main.go

run="."
dir="./..."
short="-short"
.PHONY: test
test:
	@go test --timeout=40s $(short) $(dir) -run $(run);

