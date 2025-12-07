# ==============================================================================
# Makefile helper functions for tools
#

TOOLS ?= golangci-lint goimports protoc-plugins swagger addlicense protoc-go-inject-tag protolint

.PHONY: tools.verify
tools.verify: $(addprefix tools.verify., $(TOOLS))

.PHONY: tools.install
tools.install: $(addprefix tools.install., $(TOOLS))

.PHONY: tools.install.%
tools.install.%:
	@echo "===========> Installing $*"
	@$(MAKE) install.$*

.PHONY: tools.verify.%
tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi

.PHONY: install.golangci-lint
install.golangci-lint:
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
	@golangci-lint completion bash > $(HOME)/.golangci-lint.bash
	@if ! grep -q .golangci-lint.bash $(HOME)/.bashrc; then echo "source \$$HOME/.golangci-lint.bash" >> $(HOME)/.bashrc; fi
	@golangci-lint completion zsh > $(HOME)/.golangci-lint.zsh
	@if ! grep -q .golangci-lint.zsh $(HOME)/.zshrc; then echo "source \$$HOME/.golangci-lint.zsh" >> $(HOME)/.zshrc; fi

install.goimports:
	@$(GO) install golang.org/x/tools/cmd/goimports@latest

install.protoc-plugins:
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
	@$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.0
	@$(GO) install github.com/onexstack/protoc-gen-defaults@v0.0.2
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.3
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.3

install.swagger:
	@$(GO) install github.com/go-swagger/go-swagger/cmd/swagger@latest

install.addlicense:
	@$(GO) install github.com/marmotedu/addlicense@latest

install.protoc-go-inject-tag:
	@$(GO) install github.com/favadi/protoc-go-inject-tag@latest

install.protolint:
	@$(GO) install github.com/yoheimuta/protolint/cmd/protolint@latest

# 伪目标（防止文件与目标名称冲突）
.PHONY: tools.verify tools.install tools.install.% tools.verify.% install.golangci-lint \
	install.goimports install.protoc-plugins install.swagger \
	install.addlicense install.protoc-go-inject-tag protolint
