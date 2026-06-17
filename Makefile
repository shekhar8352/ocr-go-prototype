.PHONY: test test-race bench vet lint example

test:
	go test ./... -count=1

test-race:
	go test ./... -race -count=1

bench:
	go test -bench=. -benchmem ./...

vet:
	go vet ./...

example:
	go run examples/basic/main.go
