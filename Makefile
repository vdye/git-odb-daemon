# Default target
build:

# Helpful paths
BINDIR := $(CURDIR)/bin

# Platform information
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# General targets
.PHONY: FORCE

.PHONY: build
build:
	$(RM) -r $(BINDIR)
	@mkdir -p $(BINDIR)
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build -o $(BINDIR) -ldflags "$(LDFLAGS)" ./...

.PHONY: vet
vet:
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go vet ./...

# Testing targets
.PHONY: test
test: build
	@echo "======== Running unit tests ========"
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go test ./...

# Cleanup targets
.PHONY: clean
clean:
	go clean ./...
	$(RM) -r $(BINDIR)
