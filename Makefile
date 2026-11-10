.PHONY: all build test clean run docker

BINARY_NAME=planforge

all: build test

build-frontend:
	cd web && npm install && npm run build

build-backend:
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/planforge

build: build-frontend build-backend

test:
	go test -v ./...

run: build
	./$(BINARY_NAME) serve --port 8080 --open

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf web/dist

docker:
	docker build -t planforge:latest .
