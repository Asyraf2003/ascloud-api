TOOLS_BIN := $(CURDIR)/.tools/bin

GOVULNCHECK ?= $(TOOLS_BIN)/govulncheck
GOSEC       ?= $(TOOLS_BIN)/gosec
GITLEAKS    ?= gitleaks

sec-tools:
	@mkdir -p $(TOOLS_BIN)
	GOBIN=$(TOOLS_BIN) go install golang.org/x/vuln/cmd/govulncheck@latest
	GOBIN=$(TOOLS_BIN) go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "Install gitleaks via package manager"

sec-gitleaks:
	@command -v $(GITLEAKS) >/dev/null 2>&1 || (echo "missing gitleaks. Install it first." && exit 1)
	$(GITLEAKS) detect --source . --no-banner --redact
