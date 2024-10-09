GOLANGCI_LINT_VERSION=v1.61.0

.PHONY: lint
lint: golangci-lint
	@golangci-lint run --timeout 10m0s

.PHONY: golangci-lint
golangci-lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found, installing..."; \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}; \
	else \
		INSTALLED_VERSION=$$(golangci-lint --version | awk '{print $$4}'); \
		if [ "v$$INSTALLED_VERSION" != "${GOLANGCI_LINT_VERSION}" ]; then \
			echo "updating golangci-lint from $$INSTALLED_VERSION to ${GOLANGCI_LINT_VERSION}..."; \
			$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}; \
 		fi; \
	fi