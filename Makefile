.PHONY: all
all:
	@more $(MAKEFILE_LIST)

.PHONY: clean
clean:
	$(RM) -r result

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	nix fmt

.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	nix build

.PHONY: ci
ci: lint
	nix flake check --all-systems
