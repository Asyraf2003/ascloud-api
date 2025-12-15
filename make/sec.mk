.PHONY: sec sec-vuln sec-gosec sec-gitleaks sec-tools

# Direktori terisolasi untuk binary tools
TOOLS_BIN := $(CURDIR)/.tools/bin

# Definisikan path tools, menunjuk ke lokasi terisolasi
GOVULNCHECK ?= $(TOOLS_BIN)/govulncheck
GOSEC       ?= $(TOOLS_BIN)/gosec
GITLEAKS    ?= gitleaks

# Target utama: Menjalankan semua pemeriksaan
sec: sec-vuln sec-gosec sec-gitleaks

# Target untuk menginstal semua alat keamanan
sec-tools:
	@mkdir -p $(TOOLS_BIN)
	@echo "--- Installing Go Security Tools to $(TOOLS_BIN) ---"
	# Instal govulncheck ke direktori TOOLS_BIN
	GOBIN=$(TOOLS_BIN) go install golang.org/x/vuln/cmd/govulncheck@latest
	# Instal gosec ke direktori TOOLS_BIN
	GOBIN=$(TOOLS_BIN) go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "--- Installation Complete ---"
	@echo "Untuk GITLEAKS, instal via package manager (e.g., sudo pacman -S gitleaks)"

# 1. Pemeriksaan Kerentanan Dependensi Go (Vulnerabilities)
sec-vuln:
	@echo "--- Running GOVULNCHECK ---"
	# Periksa apakah tool terinstal di lokasi yang benar, jika tidak, minta user menjalankan sec-tools
	@command -v $(GOVULNCHECK) >/dev/null 2>&1 || (echo "missing govulncheck. Run: make sec-tools" && exit 1)
	$(GOVULNCHECK) ./...

# 2. Pemeriksaan Analisis Statis Kode Go (Static Analysis)
sec-gosec:
	@echo "--- Running GOSEC ---"
	# Periksa apakah tool terinstal di lokasi yang benar
	@command -v $(GOSEC) >/dev/null 2>&1 || (echo "missing gosec. Run: make sec-tools" && exit 1)
	$(GOSEC) ./...

# 3. Pemeriksaan Kebocoran Rahasia (Secrets Detection)
sec-gitleaks:
	@echo "--- Running GITLEAKS ---"
	# Gitleaks biasanya diinstal secara sistem, jadi kita periksa di $PATH
	@command -v $(GITLEAKS) >/dev/null 2>&1 || (echo "missing gitleaks. Install it first (e.g., via pacman or brew)." && exit 1)
	$(GITLEAKS) detect --source . --no-banner --redact