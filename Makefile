.PHONY: help
help: ## Show this
	@grep -E '^[0-9a-zA-Z_-]+:(.*?## .*|[a-z _0-9]+)?$$' Makefile | sort | awk 'BEGIN {FS = ":(.*## |[\t ]*)"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

NAME:=hbtsrv
BUILD_TAG:=$(shell git describe --tags)
BUILDFLAGS:="-s -w -X github.com/lzambarda/hbt/internal.Version=$(BUILD_TAG)"

.PHONY: dependencies
dependencies: ## Install dependencies requried for development operations.
	@go get -u github.com/git-chglog/git-chglog/cmd/git-chglog
	@go mod tidy


.PHONY: build
build:
	@go build -ldflags $(BUILDFLAGS) -o ./bin/$(NAME) ./main.go

run="."
dir="./..."
short="-short"
.PHONY: test
test:
	@go test --timeout=40s $(short) $(dir) -run $(run);

.PHONY: changelog
changelog: ## Update the changelog.
	@git-chglog > CHANGELOG.md
	@echo "Changelog has been updated."


.PHONY: changelog_release
changelog_release: ## Update the changelog with a release tag.
	@git-chglog --next-tag $(tag) > CHANGELOG.md
	@echo "Changelog has been updated."
