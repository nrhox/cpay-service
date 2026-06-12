GOFMT?=gofmt "-s"
BINARY_NAME_API=app
LDFLAGS=-ldflags "-s -w"
FORMAT=.exe
PREFIX=go
BUILD_COMMAND=$(PREFIX) build

.PHONY: fmt
fmt:
	@echo "Formatting & simplifying Go code..."
	@$(GOFMT) -w .
	@echo "Done!"

.PHONY: check-fmt
check-fmt:
	@echo "Checking code formatting..."
	@output=$$($(GOFMT) -l .); \
	if [ -n "$$output" ]; then \
		echo "Files not formatted:"; \
		echo "$$output"; \
		exit 1; \
	else \
		echo "All files are properly formatted"; \
	fi

.PHONY: build-prod
build-prod: fmt
	$(BUILD_COMMAND) $(LDFLAGS) -o $(BINARY_NAME_API)$(FORMAT) cmd/api/main.go

.PHONY: run-prod
run-prod: build
	./$(BINARY_NAME_API)$(FORMAT) 

.PHONY: build
build:
	$(BUILD_COMMAND) -o $(BINARY_NAME_API)$(FORMAT) cmd/api/main.go

.PHONY: run-dev
run-dev: build
	./$(BINARY_NAME_API)$(FORMAT)

.PHONY: prepare
prepare:
	$(PREFIX) run cmd/prepare/main.go