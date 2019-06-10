all: build

build:
	mkdir -p bin
	go build -o bin/ruka ./cmd/ruka

.PHONY = build
