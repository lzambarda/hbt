.PHONY: help
help: ## Show this
	@grep -E '^[0-9a-zA-Z_-]+:(.*?## .*|[a-z _0-9]+)?$$' Makefile | sort | awk 'BEGIN {FS = ":(.*## |[\t ]*)"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

NAME:=hbtsrv
BUILD_TAG:=$(shell git describe --tags)
BUILDFLAGS:="-s -w -X github.com/lzambarda/hbt/internal.Version=$(BUILD_TAG)"

.PHONY: dependencies
dependencies: ## Install dependencies requried for development operations.
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
	@go mod tidy


.PHONY: lint
lint:
	go fmt ./...
	golangci-lint run ./...


.PHONY: build
build:
	@GOOS=darwin GOARCH=amd64 go build -ldflags $(BUILDFLAGS) -o ./bin/darwin/$(NAME) ./main.go
	@GOOS=linux GOARCH=amd64 go build -ldflags $(BUILDFLAGS) -o bin/linux/$(NAME) ./main.go


.PHONY: build_assets
build_assets: build
	@mkdir -p assets
	@tar -zcvf assets/darwin-amd64-$(NAME).tgz ./bin/darwin/$(NAME)
	@tar -zcvf assets/linux-amd64-$(NAME).tgz ./bin/linux/$(NAME)

run="."
dir="./..."
short="-short"
.PHONY: test
test:
	@go test --timeout=40s $(short) $(dir) -run $(run);
