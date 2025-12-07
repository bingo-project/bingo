# Build all by default, even if it's not first
.DEFAULT_GOAL := all

# ==============================================================================
# 定义 Makefile all 伪目标，执行 `make` 时，会默认会执行 all 伪目标
.PHONY: all
all: lint format build

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
## Testing
## --------------------------------------

##@ test:

test: ## 执行单元测试.
	@$(MAKE) go.test

cover: ## 执行单元测试，并校验覆盖率阈值.
	@$(MAKE) go.cover

## --------------------------------------
## Cleanup
## --------------------------------------

##@ clean:

clean: ## 清理构建产物、临时文件等. 例如 _output 目录.
	@echo "===========> Cleaning all build output"
	@-rm -vrf $(OUTPUT_DIR)


## --------------------------------------
## Lint / Verification
## --------------------------------------

##@ lint and verify:

lint: ## 执行静态代码检查.
	@$(MAKE) go.lint

tidy: ## 自动添加/移除依赖包.
	@$(MAKE) go.tidy

format:
	@$(MAKE) go.format

## --------------------------------------
## Generate / Manifests
## --------------------------------------

##@ generate

ca: ## 生成 CA 文件.
	@$(MAKE) gen.ca

protoc: ## 编译 protobuf 文件.
	@$(MAKE) gen.protoc

## --------------------------------------
## Hack / Tools
## --------------------------------------

##@ hack/tools:

swagger:
	@$(MAKE) swagger.docker

swagger-run: ## 聚合 swagger 文档到一个 openapi.yaml 文件中.
	@$(MAKE) swagger.run

swagger-serve: ## 运行 Swagger 文档服务器.
	@$(MAKE) swagger.serve

swag:
	@$(MAKE) swag.init

add-copyright: ## 添加版权头信息.
	@$(MAKE) gen.add-copyright

help: Makefile ## 打印 Makefile help 信息.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<TARGETS> <OPTIONS>\033[0m\n\n\033[35mTargets:\033[0m\n"} /^[0-9A-Za-z._-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' Makefile #$(MAKEFILE_LIST)
	@echo -e "$$USAGE_OPTIONS"

# 伪目标（防止文件与目标名称冲突）
.PHONY: all build test cover clean lint tidy format ca protoc swagger swagger-run swagger-serve swag add-copyright help
