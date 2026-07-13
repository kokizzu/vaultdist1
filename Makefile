GO ?= go
GOVULNCHECK ?= govulncheck
CMD ?=

.PHONY: test build-plugin vulncheck verify-dependency-security run

test:
	$(GO) test ./...

build-plugin:
	mkdir -p vault-server/plugins
	CGO_ENABLED=0 $(GO) build -o vault-server/plugins/vault-plugin-dnskey cmd/vault-plugin-dnskey/main.go

vulncheck:
	$(GOVULNCHECK) ./...

verify-dependency-security: vulncheck

run:
	@test -n "$(CMD)" || (echo "usage: make run CMD='go test ./...'" >&2; exit 2)
	$(CMD)
