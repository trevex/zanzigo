SHELL := bash

GO            ?= go
GOLANGCI_LINT ?= golangci-lint
BUF           ?= buf
FIND          ?= find

# Binaries to build
CMD_ZANZIGO = zanzigo
CMD_ZANZIGO_SRC = cmd/zanzigo/main.go

.EXPORT_ALL_VARIABLES:
.PHONY: build test bench fmt vet lint outdated protoc buf

build: $(CMD_ZANZIGO)

$(CMD_ZANZIGO): $(shell $(FIND) . -type f -name '*.go')
	mkdir -p build
	$(GO) build -o build/$(CMD_ZANZIGO) -a $(BUILDFLAGS) $(LDFLAGS) $(CMD_ZANZIGO_SRC)


test: fmt vet
	$(GO) test -v ./...

bench:
	$(GO) test -bench=. -benchmem -v ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	$(BUF) lint
	$(GO) mod verify
	$(GOLANGCI_LINT) run -v --no-config --deadline=5m

outdated:
	$(GO) list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all 2> /dev/null
	@echo "Check flake.nix and update lock-file as well!"

protoc: buf
buf:
	$(BUF) lint
	$(BUF) generate

