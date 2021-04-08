NAME := hbtsrv
BUILDFLAGS := '-s -w'

.PHONY: build
build:
	go build -ldflags $(BUILDFLAGS) -o ./bin/$(NAME) ./main.go

run="."
dir="./..."
short="-short"
.PHONY: test
test:
	@go test --timeout=40s $(short) $(dir) -run $(run);
