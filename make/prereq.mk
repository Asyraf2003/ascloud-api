.PHONY: prereq prereq-core prereq-docker

prereq-core:
	@command -v bash >/dev/null || (echo "FAIL: bash missing. Install: sudo pacman -S --needed bash"; exit 1)
	@command -v rg   >/dev/null || (echo "FAIL: rg missing. Install ripgrep"; exit 1)
	@command -v go   >/dev/null || (echo "FAIL: go missing."; exit 1)
	@echo "OK: prereq-core passed"

prereq-docker:
	@command -v docker >/dev/null || (echo "FAIL: docker missing."; exit 1)
	@echo "OK: prereq-docker passed"

prereq: prereq-core prereq-docker
	@echo "OK: prereq passed"
