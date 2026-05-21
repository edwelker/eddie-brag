.PHONY: test coverage coverage-html coverage-report

test:
	go test ./...

coverage:
	@go test ./... -coverprofile=/tmp/coverage.out
	@go tool cover -func=/tmp/coverage.out | grep total

coverage-html:
	@go test ./... -coverprofile=/tmp/coverage.out
	@go tool cover -html=/tmp/coverage.out

coverage-report:
	@go test ./... -coverprofile=/tmp/coverage.out
	@go tool cover -func=/tmp/coverage.out

build:
	go build -o ~/go/bin/brag ./cmd/brag/

install: build
	@echo "Installed to ~/go/bin/brag"
