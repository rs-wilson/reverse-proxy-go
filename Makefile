.PHONY: all build vet test clean run

all: clean build vet test

build: clean
		go build -o pom-server cmd/main.go

vet:
		go vet ./...

test: build
		source .env && go test ./...

clean:
		rm -f pom-server

run: build
	source .env && ./pom-server
