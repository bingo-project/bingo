# ==============================================================================
# Makefile helper functions for tools
#

TOOLS ?= golangci-lint goimports gotests mockgen protoc-gen-go swagger addlicense

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
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
	@golangci-lint completion bash > $(HOME)/.golangci-lint.bash
	@if ! grep -q .golangci-lint.bash $(HOME)/.bashrc; then echo "source \$$HOME/.golangci-lint.bash" >> $(HOME)/.bashrc; fi
	@golangci-lint completion zsh > $(HOME)/.golangci-lint.zsh
	@if ! grep -q .golangci-lint.zsh $(HOME)/.zshrc; then echo "source \$$HOME/.golangci-lint.zsh" >> $(HOME)/.zshrc; fi

.PHONY: install.goimports
install.goimports:
	@$(GO) install golang.org/x/tools/cmd/goimports@latest

.PHONY: install.gotests
install.gotests:
	@$(GO) install github.com/cweill/gotests/gotests@latest

.PHONY: install.mockgen
install.mockgen:
	@$(GO) install github.com/golang/mock/mockgen@latest

.PHONY: install.protoc-gen-go
install.protoc-gen-go:
	@$(GO) install github.com/golang/protobuf/protoc-gen-go@latest

.PHONY: install.swagger
install.swagger:
	@$(GO) install github.com/go-swagger/go-swagger/cmd/swagger@latest

.PHONY: install.swag
install.swag:
	@$(GO) install github.com/swaggo/swag/cmd/swag@latest

.PHONY: install.addlicense
install.addlicense:
	@$(GO) install github.com/marmotedu/addlicense@latest
