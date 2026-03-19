BINARY     := goradarr
BUILD_DIR  := bin
CMD        := ./cmd/goradarr
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")
LDFLAGS    := -s -w -X main.version=$(VERSION)

.PHONY: all build run test lint clean release fmt tidy frontend dev

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD)

run: build
	./$(BUILD_DIR)/$(BINARY)

dev:
	@which air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	air

test:
	go test ./... -race -cover

lint:
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

fmt:
	gofmt -s -w .
	goimports -w . 2>/dev/null || true

tidy:
	go mod tidy

frontend:
	cd frontend && npm install && npm run build

release: tidy frontend
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64  go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 $(CMD)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64  go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 $(CMD)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64  go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 $(CMD)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64  go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 $(CMD)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64  go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe $(CMD)

clean:
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist

docker:
	docker build -t goradarr:$(VERSION) .
