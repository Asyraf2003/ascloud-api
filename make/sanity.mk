.PHONY: sanity
sanity:
	curl -sS -D- http://localhost:8080/health -o /dev/null | rg -n "HTTP/|Content-Type:|X-Request-Id"
	curl -sS -D- http://localhost:8080/v1/health -o /dev/null | rg -n "HTTP/|Content-Type:|X-Request-Id"
	curl -sS -D- http://localhost:8080/ga-ada -o /dev/null | rg -n "HTTP/|Content-Type:|X-Request-Id"
