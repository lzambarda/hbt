NAME := hbtsrv
BUILDFLAGS := '-s -w'

.PHONY: build
build:
	go build -ldflags $(BUILDFLAGS) -o ./bin/$(NAME) ./main.go
