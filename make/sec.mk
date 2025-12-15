GOVULNCHECK ?= govulncheck
GOSEC ?= gosec
GITLEAKS ?= gitleaks

.PHONY: sec sec-vuln sec-gosec sec-gitleaks sec-tools

sec: sec-vuln sec-gosec sec-gitleaks

sec-tools:
	@echo "Install tools:"
	@echo "  go install golang.org/x/vuln/cmd/govulncheck@latest"
	@echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"
	@echo "  install gitleaks binary (package manager)"

sec-vuln:
	@command -v $(GOVULNCHECK) >/dev/null 2>&1 || (echo "missing govulncheck. Run: make sec-tools" && exit 1)
	$(GOVULNCHECK) ./...

sec-gosec:
	@command -v $(GOSEC) >/dev/null 2>&1 || (echo "missing gosec. Run: make sec-tools" && exit 1)
	$(GOSEC) ./...

sec-gitleaks:
	@command -v $(GITLEAKS) >/dev/null 2>&1 || (echo "missing gitleaks. Install it first." && exit 1)
	$(GITLEAKS) detect --no-git --source . --no-banner --redact --config .gitleaks.toml

