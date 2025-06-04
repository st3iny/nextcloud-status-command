nsc: cmd/nsc/main.go $(shell find internal -name "*.go" -type f) go.mod go.sum
	go build -o $@ $<

test:
	go test ./...

.PHONY: test
