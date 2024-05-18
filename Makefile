# Build all by default, even if it's not first
.DEFAULT_GOAL := all

# ==============================================================================
# 定义 Makefile all 伪目标，执行 `make` 时，会默认会执行 all 伪目标
.PHONY: all
all: lint format build

# ==============================================================================
# Build options

# 定义包名
ROOT_PACKAGE=bingo

# 指定应用使用的 version 包，会通过 `-ldflags -X` 向该包中指定的变量注入值
VERSION_PACKAGE=github.com/bingo-project/component-base/version


# ==============================================================================
# Includes

# make sure include common.mk at the first include line
include scripts/make-rules/common.mk
include scripts/make-rules/golang.mk
include scripts/make-rules/image.mk
include scripts/make-rules/tools.mk
include scripts/make-rules/generate.mk
include scripts/make-rules/swagger.mk

# ==============================================================================
# Usage

define USAGE_OPTIONS

Options:
  BINS             The binaries to build. Default is all of cmd.
                   This option is available when using: make build/build.multiarch
                   Example: make build BINS="bingoctl test"
  VERSION          The version information compiled into binaries.
                   The default is obtained from gsemver or git.
  V                Set to 1 enable verbose build. Default is 0.
endef
export USAGE_OPTIONS

## --------------------------------------
## Generate / Manifests
## --------------------------------------

## ca: Generate CA files for all iam components.
.PHONY: ca
ca:
	@$(MAKE) gen.ca

.PHONY: protoc
protoc: ## 编译 protobuf 文件.
	@$(MAKE) gen.protoc

.PHONY: deps
deps: ## 安装依赖，例如：生成需要的代码、安装需要的工具等.
	@$(MAKE) gen.deps

## --------------------------------------
## Binaries
## --------------------------------------

## build: Build source code for host platform.
.PHONY: build
build: tidy protoc
	@$(MAKE) go.build

.PHONY: image
image: ## Build docker images for host arch.
	@$(MAKE) image.build

## --------------------------------------
## Cleanup
## --------------------------------------

##@ clean:

.PHONY: clean
clean: ## 清理构建产物、临时文件等.
	@echo "===========> Cleaning all build output"
	@-rm -vrf $(OUTPUT_DIR)


## --------------------------------------
## Lint / Verification
## --------------------------------------

##@ lint and verify:

.PHONY: lint
lint:
	@$(MAKE) go.lint


## --------------------------------------
## Testing
## --------------------------------------

##@ test:

.PHONY: test
test: protoc ## 执行单元测试.
	@$(MAKE) go.test

.PHONY: cover
cover: protoc ## 执行单元测试，并校验覆盖率阈值.
	@$(MAKE) go.cover


## --------------------------------------
## Hack / Tools
## --------------------------------------

##@ hack/tools:

.PHONY: format
format:
	@$(MAKE) go.format

.PHONY: swagger
swagger:
	@$(MAKE) swagger.docker

.PHONY: swag
swag:
	@$(MAKE) swag.init

.PHONY: tidy
tidy:
	@$(MAKE) go.tidy

.PHONY: help
help: Makefile ## help: Show this help info.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<TARGETS> <OPTIONS>\033[0m\n\n\033[35mTargets:\033[0m\n"} /^[0-9A-Za-z._-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' Makefile #$(MAKEFILE_LIST)
	@echo -e "$$USAGE_OPTIONS"
