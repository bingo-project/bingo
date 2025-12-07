# ==============================================================================
# Makefile helper functions for swagger
#
# API_HOST 用于指定 Swagger 请求的目标地址，支持多环境切换：
#   make swagger.docker                              # 开发环境 (默认 localhost:8080)
#   make swagger.docker API_HOST=api-staging.example.com  # 预发环境
#   make swagger.docker API_HOST=api.example.com         # 生产环境
#

API_PORT := 65534
API_HOST ?= localhost:8080
SWAGGER_FILE := $(PROJ_ROOT_DIR)/api/openapi/apiserver/v1/apiserver.swagger.yaml

swagger.run: tools.verify.swagger
	@echo "===========> Generating swagger API docs"
	@swagger mixin `find $(PROJ_ROOT_DIR)/api/openapi -name "*.swagger.yaml"` \
		-q                                                    \
		--keep-spec-order                                     \
		--format=yaml                                         \
		--ignore-conflicts                                    \
		-o $(PROJ_ROOT_DIR)/api/openapi/apiserver/v1/openapi.yaml
	@echo "Generated at: $(PROJ_ROOT_DIR)/api/openapi/apiserver/v1/openapi.yaml"

swagger.serve: tools.verify.swagger
	@swagger serve -F=swagger --no-open --port 65534 $(SWAGGER_FILE) # -F=swagger or redoc

swagger.docker: ## 通过 docker 启动 swagger 文档（监听端口：65534）. 用法: make swagger.docker API_HOST=api.example.com
	@docker rm swaggerui -f 2>/dev/null || true
	@mkdir -p $(PROJ_ROOT_DIR)/_output/swagger
	@awk 'NR==11{print "host: $(API_HOST)"}1' $(SWAGGER_FILE) > $(PROJ_ROOT_DIR)/_output/swagger/apiserver.swagger.yaml
	@docker run -d --rm --name swaggerui \
       -p $(API_PORT):8080 \
       -v $(PROJ_ROOT_DIR)/_output/swagger:/tmp \
       -e SWAGGER_JSON=/tmp/apiserver.swagger.yaml \
       -e PERSIST_AUTHORIZATION=true \
       swaggerapi/swagger-ui
	@echo "Swagger UI: http://localhost:$(API_PORT)"
	@echo "API Host: $(API_HOST)"

swag.init: tools.verify.swag ## 生成 swag 文档
	@#swag fmt -d ./ --exclude ./vendor
	@swag init -d ./internal/apiserver -g router/swag.go -o ./api/swagger/apiserver -pd --parseGoList --parseInternal
	@swag init -d ./internal/admserver -g router/swag.go -o ./api/swagger/admserver -pd --parseGoList --parseInternal

# 伪目标（防止文件与目标名称冲突）
.PHONY: swagger.run swagger.serve swagger.docker swag.init
