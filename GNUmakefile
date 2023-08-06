default: testacc

SWEEP_DIR?=./internal/provider

.PHONY: sweep
sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts.";
	DC_PROJECT_ID=${DC_PROJECT_ID} DC_AUTHKEY=$(shell pwd)/authorized_key.json go test $(SWEEP_DIR) -tags=sweep -sweep=global -v -sweep-run=$(SWEEP_RUN) -timeout 60m

.PHONY: build
build:
	go build

.PHONY: testacc
test:
	go test $(TEST) -timeout=30s -parallel=4

# Run acceptance tests
.PHONY: testacc
testacc:
	DC_NETWORK_ID=${DC_NETWORK_ID} DC_PROJECT_ID=${DC_PROJECT_ID} DC_AUTHKEY=$(shell pwd)/authorized_key.json TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: lint
lint: tools
	@echo "==> Checking source code against linters..."
	golangci-lint run -c .golangci.yml ./$(PKG_NAME)/...

.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: doc-tools
docs: doc-tools
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate