nsc: cmd/nsc/main.go $(shell find internal -name "*.go" -type f)
	go build -o $@ $<

test:
	go test ./...

.PHONY: test
