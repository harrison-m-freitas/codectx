# Makefile — codectx
# Reqs: GNU make, bash, tar, sha256sum (ou shasum), go >= 1.22
# Use:   make help
#        make test | make fmt | make build
#        make cross | make pack | make checksums | make verify
#        make release-local
#        make sha-native

SHELL := /usr/bin/env bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

.PHONY: help clean fmt test build sha-native cross pack checksums verify release-local install

# Verbose toggle: make VERBOSE=1 …
VERBOSE ?= 0
ifeq ($(VERBOSE),1)
Q :=
else
Q := @
endif

# ------------------------------------------------------------------------------
# Config
# ------------------------------------------------------------------------------
BIN         := codectx
MODULE      := ./cmd/codectx
GO          ?= go
CGO_ENABLED ?= 0
GOFLAGS     ?= -trimpath
LDFLAGS     ?= -s -w -buildid=
DIST        := dist

# versão vem da última tag; fallback se não houver
VERSION   ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0-dev)
COMMIT      := $(shell git rev-parse --short=8 HEAD 2>/dev/null || echo unknown)
DATE        := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
DIST_DIR    := $(DIST)/$(VERSION)

# Matriz de cross build (ajuste se quiser)
OS_ARCH := linux/amd64 darwin/arm64 windows/amd64

# Descobertas do host para build "nativo"
GOOS_LOCAL   := $(shell $(GO) env GOOS)
GOARCH_LOCAL := $(shell $(GO) env GOARCH)
EXE_EXT      := $(shell [ "$(GOOS_LOCAL)" = windows ] && printf ".exe")
BIN_NATIVE_DIR  := $(DIST_DIR)/$(BIN)_$(GOOS_LOCAL)_$(GOARCH_LOCAL)
BIN_NATIVE_FILE := $(BIN_NATIVE_DIR)/$(BIN)$(EXE_EXT)

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------
##@ Utilitários
## Mostra ajuda com alvos e descrições
help:
	@echo "Targets:";
	@awk 'BEGIN{FS=":.*##"; OFS=""; sec=""} \
		/^##@/ {sec=substr($$0,5); printf "\n\033[1m%s\033[0m\n", sec; next} \
		/^[a-zA-Z0-9_%-]+:.*##/ {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' \
		$(MAKEFILE_LIST); echo

## Remove diretório dist/
clean: ## Limpa artefatos (dist/)
	$(Q)rm -rf "$(DIST)"

## Formata com go fmt
fmt: ## go fmt ./...
	$(Q)$(GO) fmt ./...

## Roda a suíte de testes
test: ## go test ./...
	$(Q)$(GO) test ./...

# ------------------------------------------------------------------------------
# Build
# ------------------------------------------------------------------------------
## Build nativo (SO/arch local)
build: ## Build nativo (gera $(BIN_NATIVE_FILE))
	$(Q)mkdir -p "$(BIN_NATIVE_DIR)"
	@printf ">> Building native: %s/%s\n" "$(GOOS_LOCAL)" "$(GOARCH_LOCAL)"
	$(Q)CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' \
		-o "$(BIN_NATIVE_FILE)" $(MODULE)

## Mostra sha256 do binário nativo (depois de build)
sha-native: build ## sha256 do binário nativo
	@echo ">> sha256sum $(BIN_NATIVE_FILE)"
	@{ command -v sha256sum >/dev/null 2>&1 && sha256sum "$(BIN_NATIVE_FILE)"; } || \
	{ command -v shasum >/dev/null 2>&1 && shasum -a 256 "$(BIN_NATIVE_FILE)"; }

# ------------------------------------------------------------------------------
# Cross / Pack
# ------------------------------------------------------------------------------
## Cross-compile para OS/ARCH (matriz)
cross: ## Cross build (linux/amd64, darwin/arm64, windows/amd64)
	$(Q)mkdir -p "$(DIST_DIR)"
	@for pair in $(OS_ARCH); do \
		GOOS=$${pair%/*}; GOARCH=$${pair#*/}; \
		outdir="$(DIST_DIR)/$(BIN)_$${GOOS}_$${GOARCH}"; \
		binname="$(BIN)"; \
		if [ "$$GOOS" = windows ]; then binname="$${binname}.exe"; fi; \
		printf ">> Building %s/%s → %s/%s\n" "$$GOOS" "$$GOARCH" "$$outdir" "$$binname"; \
		mkdir -p "$$outdir"; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS="$$GOOS" GOARCH="$$GOARCH" \
		  $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' \
		  -o "$$outdir/$$binname" $(MODULE); \
	done

## Empacota cada pasta (tar.gz) com versão no nome
pack: cross ## Empacota em .tar.gz (versionado)
	@cd "$(DIST_DIR)" && \
	for d in $(BIN)_*; do \
		[ -d "$$d" ] || continue; \
		printf ">> Packing %s → %s-%s.tar.gz\n" "$$d" "$$d" "$(VERSION)"; \
		tar -czf "$$d-$(VERSION).tar.gz" "$$d"; \
 	done

# ------------------------------------------------------------------------------
# Checksums
# ------------------------------------------------------------------------------
## Gera SHA256SUMS.txt e SHA256SUMS.json (arquivo → sha)
## Gera SHA256SUMS.txt e SHA256SUMS.json (arquivo → sha)
checksums: pack ## Gera checksums e manifesto
	@cd "$(DIST_DIR)" && \
	echo ">> Generating SHA256SUMS.txt" && \
	{ command -v sha256sum >/dev/null 2>&1 && sha256sum *.tar.gz > SHA256SUMS.txt; } || \
	{ command -v shasum   >/dev/null 2>&1 && shasum -a 256 *.tar.gz > SHA256SUMS.txt; } && \
	echo ">> Generating SHA256SUMS.json" && \
	awk 'BEGIN { print "{" } \
	     { hash=$$1; $$1=""; sub(/^  /,""); file=$$0; gsub(/"/,"\\\"",file); \
	       lines[NR]=sprintf("  \"%s\": \"%s\"", file, hash) } \
	     END { for (i=1; i<=NR; i++) { \
	              printf "%s%s\n", lines[i], (i<NR ? "," : "") } \
	           print "}" }' SHA256SUMS.txt > SHA256SUMS.json && \
	echo && echo "==> SHA256SUMS.txt" && cat SHA256SUMS.txt


## Verifica checksums
verify: ## sha256sum -c SHA256SUMS.txt
	@cd "$(DIST_DIR)" && \
	echo ">> Verifying checksums" && \
	sha256sum -c SHA256SUMS.txt 2>/dev/null || shasum -a 256 -c SHA256SUMS.txt

# ------------------------------------------------------------------------------
# Release local
# ------------------------------------------------------------------------------
## Limpa → testa → cross → pack → checksums (artefatos em dist/<versão>)
release-local: clean test checksums ## Pipeline de release local
	@echo; echo "Artifacts em: $(DIST_DIR)"; ls -1 "$(DIST_DIR)"

# ------------------------------------------------------------------------------
# Instalação
# ------------------------------------------------------------------------------
## Instala binário nativo em ~/.local/bin
install: build ## Instala em ~/.local/bin
	$(Q)install -d "$$HOME/.local/bin"
	$(Q)install -m 0755 "$(BIN_NATIVE_FILE)" "$$HOME/.local/bin/$(BIN)"
	@echo ">> Installed to $$HOME/.local/bin/$(BIN)"
